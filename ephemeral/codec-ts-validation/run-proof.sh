#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 3 ]]; then
  echo "usage: $0 CANDIDATE_PATH CANDIDATE_SHA OUTPUT_DIR" >&2
  exit 2
fi

candidate_path=$(cd "$1" && pwd -P)
candidate_sha=$2
output_dir=$3
fixture_dir=$(cd "$(dirname "$0")/consumer" && pwd -P)
actual_sha=$(git -C "$candidate_path" rev-parse HEAD)
if [[ "$actual_sha" != "$candidate_sha" ]]; then
  echo "candidate HEAD $actual_sha does not match requested $candidate_sha" >&2
  exit 2
fi
if ! git -C "$candidate_path" diff --quiet || ! git -C "$candidate_path" diff --cached --quiet; then
  echo "candidate has tracked changes; commit an exact reviewable candidate before proof" >&2
  exit 2
fi
while IFS= read -r -d '' untracked; do
  case "$untracked" in
    ephemeral/*) ;;
    *)
      echo "candidate has untracked non-ephemeral input $untracked; commit or remove it before proof" >&2
      exit 2
      ;;
  esac
done < <(git -C "$candidate_path" ls-files --others --exclude-standard -z)

mkdir -p "$output_dir"
output_dir=$(cd "$output_dir" && pwd -P)
run_dir=$(mktemp -d "${TMPDIR:-/tmp}/codec-ts-consumer.XXXXXX")
cleanup() {
  python3 -c 'import shutil, sys; shutil.rmtree(sys.argv[1])' "$run_dir"
}
trap cleanup EXIT
consumer_dir="$run_dir/consumer"
mkdir -p "$consumer_dir" "$run_dir/bin"
cp -R "$fixture_dir"/. "$consumer_dir"/

canonical_hash_before=$(find "$fixture_dir" -type f -print0 | sort -z | xargs -0 shasum | shasum | awk '{print $1}')
candidate_state_before=$(git -C "$candidate_path" status --porcelain=v1 --untracked-files=all | shasum | awk '{print $1}')

{
  echo "candidate_path=$candidate_path"
  echo "candidate_sha=$candidate_sha"
  echo "go_version=$(go version)"
  echo "node_version=$(node --version)"
  echo "npm_version=$(npm --version)"
  echo "canonical_fixture_sha1=$canonical_hash_before"
} > "$output_dir/provenance.txt"
git -C "$candidate_path" status --porcelain=v1 --untracked-files=all > "$output_dir/candidate-status.txt"

(
  cd "$candidate_path"
  GOFLAGS=-p=2 go build -trimpath -o "$run_dir/bin/gen-jsonschema" ./gen-jsonschema
)
binary_sha1=$(shasum "$run_dir/bin/gen-jsonschema" | awk '{print $1}')
echo "$binary_sha1  gen-jsonschema" > "$output_dir/generator-binary.sha1"

(
  cd "$consumer_dir"
  go mod edit -replace="github.com/tylergannon/go-gen-jsonschema=$candidate_path"
  GOFLAGS=-p=2 go mod tidy
  "$run_dir/bin/gen-jsonschema" gen \
    --target . \
    --validate \
    --formats json \
    --typescript ./generated/ts \
    --typescript-barrel \
    > "$output_dir/generate.stdout" \
    2> "$output_dir/generate.stderr"
  # Generated validation code introduces its validator dependency, which is not
  # visible to the pre-generation tidy used to build the input package.
  GOFLAGS=-p=2 go mod tidy
  GOFLAGS=-p=2 go test -count=1 ./... -v \
    > "$output_dir/go-test.stdout" \
    2> "$output_dir/go-test.stderr"
  npm ci --ignore-scripts \
    > "$output_dir/npm-ci.stdout" \
    2> "$output_dir/npm-ci.stderr"
  ./node_modules/.bin/tsc --project tsconfig.json --pretty false \
    > "$output_dir/tsc.stdout" \
    2> "$output_dir/tsc.stderr"
)

mkdir -p "$output_dir/generated/ts" "$output_dir/generated/jsonschema"
cp "$consumer_dir/jsonschema_gen.go" "$output_dir/generated/jsonschema_gen.go"
cp "$consumer_dir/generated/ts/types.ts" "$consumer_dir/generated/ts/index.ts" "$output_dir/generated/ts/"
cp "$consumer_dir/jsonschema"/* "$output_dir/generated/jsonschema/"
cp "$consumer_dir/go.mod" "$consumer_dir/go.sum" "$output_dir/generated/"

canonical_hash_after=$(find "$fixture_dir" -type f -print0 | sort -z | xargs -0 shasum | shasum | awk '{print $1}')
candidate_state_after=$(git -C "$candidate_path" status --porcelain=v1 --untracked-files=all | shasum | awk '{print $1}')
if [[ "$canonical_hash_before" != "$canonical_hash_after" ]]; then
  echo "canonical consumer fixture changed during proof" >&2
  exit 1
fi
if [[ "$candidate_state_before" != "$candidate_state_after" ]]; then
  echo "candidate tracked state changed during proof" >&2
  exit 1
fi

{
  echo "candidate_sha=$candidate_sha"
  echo "generator_exit=0"
  echo "go_test_exit=0"
  echo "typescript_compile_exit=0"
  echo "canonical_fixture_unchanged=true"
  echo "candidate_tracked_state_unchanged=true"
} > "$output_dir/result.txt"

(
  cd "$output_dir"
  find . -type f ! -name artifact-sha1.txt -print0 | sort -z | xargs -0 shasum
) > "$output_dir/artifact-sha1.txt"

cat "$output_dir/result.txt"
