# Adversarial review: issue #60 agent CLI

Reviewed the complete `origin/main...ea77d628faf584fce3e528012d11a17897008c08` diff in the issue-60 worktree, the surrounding builder/syntax code, the accepted v1 contract, issue #60 acceptance text, and the fresh proof under `ephemeral/issue-60/`. The current branch also contains the later proof-only commit `383c7e2`; the parent’s uncommitted documentation alignment was checked separately. No product files were changed by this review.

## Verdict

**Not merge-ready.** The fresh consumer proof demonstrates the happy path and several negative cases, but three implementation defects still violate the machine inspection contract. The first two are P1 because they either corrupt machine stdout or report a valid model as a toolchain failure; the third can produce a false supported result for active production hooks.

## Findings

### [P1] Inspection can write a Go stack/debug line before the JSON result

`internal/builder/gen_schema.go:1227` still calls `fmt.Println(string(debug.Stack()))`. The call is in `resolveEmbeddedType`, which is reached from `renderStructProps` (`:1232-1245`) while `builder.Inspect` maps a registered root. A valid embedded selector such as `remote.Type` (or another unsupported embedded expression) can reach this default branch. `renderSchema` also retains a direct `fmt.Printf` debug path at `internal/builder/gen_schema.go:912`.

The CLI’s machine path calls `inspection.Inspect` and then emits its JSON envelope, so any such diagnostic text is written directly to process stdout before/alongside the envelope. The documented contract requires stdout to contain one JSON document, and issue #60 explicitly includes removal of the #34 debug-dump behavior. A fresh agent parsing JSON therefore fails exactly on a valid-but-unsupported model, even though the structured result would otherwise be useful. All mapper/debug paths need to return typed inspection information or write nowhere on the machine path.

### [P1] Valid generic models are classified as toolchain failures

The fresh boundary probe at `v1-wave2/ephemeral/wave2/generic-inspection-consumer` uses a valid model with `type Box[T any] struct { Value T }` and `type Root struct { Box[string] }`. `internal/syntax/scan_result.go:513-519` deliberately returns a plain error for generic expressions. `internal/builder/inspection.go:72-75` wraps that scan error without a typed `ScanError` or `PackageLoadError`, and `internal/inspection/inspect.go:26-43` defaults every untyped error to classification `toolchain`, code `package_load_failed`, status `error`, and exit 1.

This is a valid Go package whose shape is outside the v1 inspection boundary. Issue #60 requires an agent to distinguish valid unsupported/unknown models (exit 3) from invalid input/toolchain failure (exit 1). The probe consequently misleads automation into repairing dependencies or the toolchain when the model itself merely needs an unsupported-shape diagnosis. Preserve a typed scan classification (or convert this case into a root-level unsupported/unknown diagnostic) before mapping it through the CLI.

### [P1] Production hook discovery does not use the same build tags as inspection

`internal/syntax/loader.go:25-29` and `LoadReadonly` (`:35-41`) always load the target with the `jsonschema` build tag. `internal/syntax/json_methods.go:74-91`, however, starts from `build.Default` and adds only `-tags` found in `GOFLAGS`; it never adds `syntax.BuildTag` or otherwise receives the active tags from the loader.

A real production `MarshalJSON`/`UnmarshalJSON` method in a file tagged `//go:build jsonschema` is active in the inspected package but is skipped by `FindProductionJSONMethods`, so a root can be reported supported despite an unproved custom wire hook. Conversely, when a caller sets `GOFLAGS=-tags=jsonschema`, the scanner starts treating the generation-only declaration methods in the build-tagged `schema.go` as production methods (it excludes only `jsonschema_gen.go`), producing false `unknown_custom_json_hook` findings. The scanner must share the exact active build configuration and retain the generation-stub exclusion guarantee.

### [P2] Published exit table contradicts the unregistered-type result

The current parent-alignment draft `docs/agent-cli.md:36-40` says exit 2 covers “invalid request, target, registration, or Go source.” The implemented unregistered request path at `internal/inspection/inspect.go:56-67` intentionally produces `type_not_registered`, `StatusUnknown`, and therefore exit 3; the explicit command contract in `ephemeral/issue-60/command-contract.md` also says exit 3 covers unsupported/unknown. Clarify the table so consumers do not treat an unregistered type name as a malformed registration request. If “registration” refers only to malformed marker source, say that explicitly.

## Evidence and checks

- `ephemeral/issue-60/proof-result.txt` records the fresh consumer matrix: supported 0, unsupported/unknown 3, invalid 2, toolchain 1, JSON-only stdout with zero stderr, and unchanged consumer files.
- The proof covers ordinary supported/unsupported/custom-hook/external/unregistered cases, but not a build-tagged production hook or a generic unsupported model, which is why those boundary defects escaped it.
- Scoped checks run during review: `GOFLAGS=-p=2 go test ./gen-jsonschema ./internal/inspection ./internal/syntax` — **pass**.
- The worker worklog records baseline and final `GOFLAGS=-p=2 go test ./...` passes; no concurrent full-suite or destructive fixture regeneration was run by this review.
