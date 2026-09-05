#!/usr/bin/env bash
set -euo pipefail

release_version=v1.0.0-rc.5
expected_tag_object=db33bee926d6fda42b447a1b28679400dc1c0342
expected_tag_commit=8d83354421e2f37ea42bfaccdcbd1ff9211715a2
tool_package=github.com/tylergannon/go-gen-jsonschema/gen-jsonschema
module_package=github.com/tylergannon/go-gen-jsonschema
proof_dir=$(cd "$(dirname "$0")" && pwd -P)
fixture_dir="$proof_dir/fixture"
evidence_dir="$proof_dir/evidence"

if [[ -e "$evidence_dir" ]]; then
  echo "evidence already exists at $evidence_dir; use a clean checkout for a fresh install" >&2
  exit 2
fi
if grep -Eq '^replace[[:space:] (]' "$fixture_dir/go.mod"; then
  echo "fixture go.mod must not contain a replace directive" >&2
  exit 2
fi
if grep -q "$module_package" "$fixture_dir/go.mod"; then
  echo "fixture go.mod must begin without the release dependency" >&2
  exit 2
fi
for command in go node npm shasum; do
  if ! command -v "$command" >/dev/null; then
    echo "required command is unavailable: $command" >&2
    exit 2
  fi
done

if [[ "${1:-}" == "--prepare-only" ]]; then
  if [[ $# -ne 1 ]]; then
    echo "usage: $0 [--prepare-only]" >&2
    exit 2
  fi
  echo "prepared_release=$release_version"
  echo "fixture_dependency=absent"
  echo "fixture_replace_directive=absent"
  echo "network_install_executed=false"
  exit 0
fi
if [[ $# -ne 0 ]]; then
  echo "usage: $0 [--prepare-only]" >&2
  exit 2
fi

run_dir=$(mktemp -d "${TMPDIR:-/tmp}/rc5-install-smoke.XXXXXX")
consumer_dir="$run_dir/consumer"
results_dir="$run_dir/results"
staged_evidence="$run_dir/evidence"
mkdir -p "$consumer_dir" "$results_dir" "$staged_evidence/consumer" "$staged_evidence/results"
cp -R "$fixture_dir"/. "$consumer_dir"/

cleanup() {
  status=$?
  if [[ $status -eq 0 ]]; then
    rm -rf "$run_dir"
  else
    echo "failed smoke retained at $run_dir" >&2
  fi
}
trap cleanup EXIT

fixture_sha256_before=$(find "$fixture_dir" -type f -print0 | sort -z | xargs -0 shasum -a 256 | shasum -a 256 | awk '{print $1}')
git -C "$proof_dir" ls-remote --tags origin "refs/tags/$release_version*" > "$results_dir/published-tag.txt"
tab=$'\t'
if ! grep -Fxq "$expected_tag_object${tab}refs/tags/$release_version" "$results_dir/published-tag.txt" || \
   ! grep -Fxq "$expected_tag_commit${tab}refs/tags/$release_version^{}" "$results_dir/published-tag.txt"; then
  echo "published $release_version tag does not resolve to expected commit $expected_tag_commit" >&2
  exit 1
fi

(
  cd "$consumer_dir"
  go get -tool "$tool_package@$release_version" \
    > "$results_dir/go-get-tool.stdout" \
    2> "$results_dir/go-get-tool.stderr"
  go mod tidy \
    > "$results_dir/pre-generate-tidy.stdout" \
    2> "$results_dir/pre-generate-tidy.stderr"

  if grep -Eq '^replace[[:space:] (]' go.mod; then
    echo "installed consumer go.mod contains a replace directive" >&2
    exit 1
  fi
  resolved_module=$(go list -m -f '{{.Path}}@{{.Version}}' "$module_package")
  if [[ "$resolved_module" != "$module_package@$release_version" ]]; then
    echo "runtime module resolved to $resolved_module, want $module_package@$release_version" >&2
    exit 1
  fi

  go tool gen-jsonschema gen \
    --target . \
    --validate \
    --formats json \
    --typescript ./generated/ts \
    --typescript-barrel \
    > "$results_dir/generate.stdout" \
    2> "$results_dir/generate.stderr"
  go mod tidy \
    > "$results_dir/post-generate-tidy.stdout" \
    2> "$results_dir/post-generate-tidy.stderr"
  go tool gen-jsonschema gen \
    --target . \
    --validate \
    --formats json \
    --typescript ./generated/ts \
    --typescript-barrel \
    --no-changes \
    > "$results_dir/regenerate.stdout" \
    2> "$results_dir/regenerate.stderr"

  npm ci --ignore-scripts --no-audit --no-fund \
    > "$results_dir/npm-ci.stdout" \
    2> "$results_dir/npm-ci.stderr"
  ./node_modules/.bin/tsc --project tsconfig.json --pretty false \
    > "$results_dir/tsc.stdout" \
    2> "$results_dir/tsc.stderr"
  go run ./cmd/bridge emit "$results_dir/go-to-node.json" \
    > "$results_dir/go-emit.stdout" \
    2> "$results_dir/go-emit.stderr"
  node ./runtime.mjs from-go "$results_dir/go-to-node.json" "$results_dir/node-to-go.json" \
    > "$results_dir/node-from-go.stdout" \
    2> "$results_dir/node-from-go.stderr"
  go run ./cmd/bridge consume "$results_dir/node-to-go.json" "$results_dir/go-back-to-node.json" \
    > "$results_dir/go-consume.stdout" \
    2> "$results_dir/go-consume.stderr"
  node ./runtime.mjs verify-back "$results_dir/go-back-to-node.json" \
    > "$results_dir/node-verify-back.stdout" \
    2> "$results_dir/node-verify-back.stderr"
  go test ./... \
    > "$results_dir/consumer-go-test.stdout" \
    2> "$results_dir/consumer-go-test.stderr"

  go tool -n gen-jsonschema > "$results_dir/tool-command.txt"
  go version -m "$(cat "$results_dir/tool-command.txt")" > "$results_dir/tool-build-info.txt"
  go mod edit -json > "$results_dir/go-mod.json"
)

fixture_sha256_after=$(find "$fixture_dir" -type f -print0 | sort -z | xargs -0 shasum -a 256 | shasum -a 256 | awk '{print $1}')
if [[ "$fixture_sha256_before" != "$fixture_sha256_after" ]]; then
  echo "canonical fixture changed during smoke" >&2
  exit 1
fi

resolved_module=$(cd "$consumer_dir" && go list -m -f '{{.Path}}@{{.Version}}' "$module_package")
typescript_version=$(node -p "require('$consumer_dir/node_modules/typescript/package.json').version")
tool_binary=$(cat "$results_dir/tool-command.txt")
tool_sha256=$(shasum -a 256 "$tool_binary" | awk '{print $1}')

{
  echo "release_version=$release_version"
  echo "release_tag_object=$expected_tag_object"
  echo "release_tag_commit=$expected_tag_commit"
  echo "tool_package=$tool_package@$release_version"
  echo "tool_invocation=go tool gen-jsonschema"
  echo "tool_binary_sha256=$tool_sha256"
  echo "consumer_module=$resolved_module"
  echo "consumer_replace_directive=absent"
  echo "go_version=$(go version)"
  echo "node_version=$(node --version)"
  echo "npm_version=$(npm --version)"
  echo "typescript_version=$typescript_version"
  echo "canonical_fixture_sha256=$fixture_sha256_before"
} > "$results_dir/provenance.txt"

{
  cat "$results_dir/go-emit.stdout"
  cat "$results_dir/node-from-go.stdout"
  cat "$results_dir/go-consume.stdout"
  cat "$results_dir/node-verify-back.stdout"
  echo "installed_tool_version_matches_runtime=true"
  echo "generation_idempotent=true"
  echo "typescript_compile=true"
  echo "consumer_go_test=true"
  echo "canonical_fixture_unchanged=true"
} > "$results_dir/result.txt"

mkdir -p "$staged_evidence/consumer/cmd/bridge"
cp "$consumer_dir"/.gitignore "$consumer_dir"/go.mod "$consumer_dir"/go.sum \
  "$consumer_dir"/package.json "$consumer_dir"/package-lock.json \
  "$consumer_dir"/runtime.mjs "$consumer_dir"/schema.go \
  "$consumer_dir"/transport.ts "$consumer_dir"/tsconfig.json \
  "$consumer_dir"/types.go "$consumer_dir"/jsonschema_gen.go \
  "$staged_evidence/consumer/"
cp "$consumer_dir"/cmd/bridge/*.go "$staged_evidence/consumer/cmd/bridge/"
cp -R "$consumer_dir"/generated "$consumer_dir"/jsonschema "$staged_evidence/consumer/"
cp -R "$results_dir"/. "$staged_evidence/results/"

(
  cd "$staged_evidence"
  find consumer results -type f ! -name artifact-sha256.txt -print0 \
    | sort -z \
    | xargs -0 shasum -a 256
) > "$staged_evidence/results/artifact-sha256.txt"

mkdir "$evidence_dir"
cp -R "$staged_evidence"/. "$evidence_dir"/

cat "$evidence_dir/results/provenance.txt"
cat "$evidence_dir/results/result.txt"
