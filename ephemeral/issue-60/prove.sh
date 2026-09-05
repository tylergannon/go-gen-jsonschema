#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)
fixture_root="$repo_root/ephemeral/issue-60/consumer"
build_root=$(mktemp -d "${TMPDIR:-/tmp}/issue-60-build.XXXXXX")
trap 'rm -rf "$build_root"' EXIT INT TERM

(
	cd "$repo_root"
	go build -o "$build_root/gen-jsonschema" ./gen-jsonschema
)

proof_root=$(mktemp -d "$repo_root/ephemeral/issue-60/run.XXXXXX")
trap 'rm -rf "$build_root" "$proof_root"' EXIT INT TERM

cp -R "$fixture_root/model" "$fixture_root/broken" "$fixture_root/invalid" "$proof_root/"

(
	cd "$proof_root/model"
	go mod edit -replace "github.com/tylergannon/go-gen-jsonschema=$repo_root"
	go mod tidy
)

proof_bin="$build_root/gen-jsonschema"
"$proof_bin" version --json

before=$(find "$proof_root/model" "$proof_root/broken" "$proof_root/invalid" -type f -print0 | sort -z | xargs -0 shasum)

run_case() {
	label=$1
	expected=$2
	shift 2
	stderr_file="$proof_root/stderr"
	set +e
	output=$("$proof_bin" "$@" 2>"$stderr_file")
	actual=$?
	set -e
	if [ "$actual" -ne "$expected" ]; then
		printf '%s: exit %s, expected %s\n' "$label" "$actual" "$expected" >&2
		exit 1
	fi
	if [ -s "$stderr_file" ]; then
		printf '%s: machine mode wrote stderr\n' "$label" >&2
		exit 1
	fi
	parsed=$(printf '%s' "$output" | jq -ce '{kind,status,usage:(.usage // null),types:[.types[]? | {typePath,status,diagnostics:[.diagnostics[]? | {code,classification,fieldPath:(.fieldPath // null),source:(.source // null)}]}],diagnostics:[.diagnostics[]? | {code,classification,source:(.source // null)}]}')
	printf '%s exit=%s stderr_bytes=0 %s\n' "$label" "$actual" "$parsed"
}

run_case supported 0 inspect --json --target "$proof_root/model" Supported
run_case unsupported 3 inspect --json --target "$proof_root/model" Unsupported
run_case unknown 3 inspect --json --target "$proof_root/model" Unknown
run_case optional_missing_omitzero 3 inspect --json --target "$proof_root/model" OptionalMissingOmitzero
run_case predeclared_any 3 inspect --json --target "$proof_root/model" AnyModel
run_case unregistered 3 inspect --json --target "$proof_root/model" Missing
run_case invalid_flag 2 inspect --json --target "$proof_root/model" --bad-flag
run_case invalid_source 2 inspect --json --target "$proof_root/invalid" Broken
run_case toolchain 1 inspect --json --target "$proof_root/broken" Broken
run_case json_help 0 version --json --help

after=$(find "$proof_root/model" "$proof_root/broken" "$proof_root/invalid" -type f -print0 | sort -z | xargs -0 shasum)
if [ "$before" != "$after" ]; then
	printf 'all_consumer_files_unchanged=false\n' >&2
	exit 1
fi
printf 'all_consumer_files_unchanged=true\n'

if [ -e "$proof_root/broken/go.sum" ]; then
	printf 'broken_go_sum_created=true\n' >&2
	exit 1
fi
printf 'broken_go_sum_created=false\n'
