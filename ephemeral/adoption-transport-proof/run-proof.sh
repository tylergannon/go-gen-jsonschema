#!/usr/bin/env bash
set -euo pipefail

proof_dir=$(cd "$(dirname "$0")" && pwd -P)
repo_dir=$(cd "$proof_dir/../.." && pwd -P)
consumer_dir="$proof_dir/consumer"
results_dir="$proof_dir/results"
run_dir=$(mktemp -d "${TMPDIR:-/tmp}/adoption-transport-proof.XXXXXX")
trap 'rm -rf "$run_dir"' EXIT

mkdir -p "$results_dir"

if grep -Eq '^replace[[:space:] (]' "$consumer_dir/go.mod"; then
  echo "consumer go.mod must not contain a replace directive" >&2
  exit 1
fi

generator_source_commit=$(git -C "$repo_dir" log -1 --format=%H -- gen-jsonschema internal go.mod go.sum)
go_version=$(go version)
node_version=$(node --version)
npm_version=$(npm --version)

(
  cd "$repo_dir"
  go build -trimpath -o "$run_dir/gen-jsonschema" ./gen-jsonschema
)
generator_sha256=$(shasum -a 256 "$run_dir/gen-jsonschema" | awk '{print $1}')

(
  cd "$consumer_dir"
  # Prime the package checksum before the generator loads the jsonschema-tagged
  # registration file. A bare module download may retain only the go.mod sum.
  go mod tidy
  "$run_dir/gen-jsonschema" gen \
    --target . \
    --validate \
    --formats json \
    --typescript ./generated/ts \
    --typescript-barrel \
    > "$results_dir/generate.stdout" \
    2> "$results_dir/generate.stderr"
  go mod tidy
  "$run_dir/gen-jsonschema" gen \
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
)

consumer_module=$(cd "$consumer_dir" && go list -m -f '{{.Path}}@{{.Version}}' github.com/tylergannon/go-gen-jsonschema)
typescript_version=$(node -p "require('$consumer_dir/node_modules/typescript/package.json').version")

{
  echo "generator_source_commit=$generator_source_commit"
  echo "generator_binary_sha256=$generator_sha256"
  echo "consumer_module=$consumer_module"
  echo "consumer_replace_directive=absent"
  echo "go_version=$go_version"
  echo "node_version=$node_version"
  echo "npm_version=$npm_version"
  echo "typescript_version=$typescript_version"
} > "$results_dir/provenance.txt"

{
  cat "$results_dir/go-emit.stdout"
  cat "$results_dir/node-from-go.stdout"
  cat "$results_dir/go-consume.stdout"
  cat "$results_dir/node-verify-back.stdout"
  echo "generation_idempotent=true"
  echo "typescript_compile=true"
  echo "consumer_go_test=true"
} > "$results_dir/result.txt"

(
  cd "$proof_dir"
  find consumer/generated consumer/jsonschema results -type f ! -name artifact-sha256.txt -print0 \
    | sort -z \
    | xargs -0 shasum -a 256
) > "$results_dir/artifact-sha256.txt"

cat "$results_dir/provenance.txt"
cat "$results_dir/result.txt"
