# Issue 77: regenerable advertised examples

decision: Work issues sequentially by user request; begin with #77, then #80.
decision: Delegate implementation and repository research to Claude through Tractor; root owns scope, review, integration, and merge.
friction: Installed Tractor CLI has no `workflow list` command even though older local guidance named it as the availability authority -> use the installed CLI help and bundled pipeline examples for this run.

## Implementation

- `examples/interfaces_options`: added `//go:generate go run ../../gen-jsonschema/`
  to `types.go` (the declarations were already valid; only the entry point and
  checked-in artifacts were missing). Ran `go generate` and checked in the
  resulting `jsonschema_gen.go` and `jsonschema/Owner.json(.sum)`.
- `examples/enums_stringmode`: `schema.go` registered an undeclared
  `TestCase.Schema` / `TestCase{}.Priority`. Rewrote the registration to use
  the package's actual `Paint`/`Color` declarations with
  `jsonschema.WithStringerEnum(Paint{}.C)`, demonstrating the intended
  numeric-enum-as-string-mode behavior without introducing new types. Added
  the same `//go:generate` directive and checked in the generated artifacts.
- Verified for both: `go generate` succeeds, and a subsequent
  `gen --no-changes` / second `go generate` run is a true no-op.
- friction: running a broad, repo-wide `go generate ./...` sweep (not scoped
  to the two fixed example directories) also regenerated
  `examples/structs/jsonschema_gen.go` with stale content (dropped
  `ValidateJSON`/compiled-schema code), because that sweep didn't pass
  `examples/structs`'s own `--validate` directive. This was not a concurrent
  process touching the worktree — it was this run's own broad,
  under-flagged sweep. Fixed by re-running `go generate` inside
  `examples/structs` (using its own `--validate` directive), which restored
  it to match the committed content exactly (confirmed via `git diff`). Left
  everything else in the repo untouched.
- Added `examples_regenerate_test.go` (root package): a compact regression
  test covering just the two examples #77 fixed. For each example it (1)
  checks for a real `//go:generate` directive invoking `gen-jsonschema`, (2)
  checks the expected generated artifacts exist on disk (`jsonschema_gen.go`,
  `jsonschema/*.json(.sum)`), and (3) proves those artifacts are current by
  running the real generator (`internal/builder.Run`, in-process, the same
  entry point the CLI's `--no-changes` flag uses) against a copy of the
  example. The copy is built as a throwaway module (its own `go.mod`/`go.sum`
  plus the root package source the example imports) at the same
  `examples/<name>` path so import-path-derived identifiers (e.g. interface
  codec helper names) regenerate identically; `--no-changes` catches any
  JSON schema drift, and the regenerated `jsonschema_gen.go` is diffed
  byte-for-byte against the checked-in one. Needs no `.git` metadata and
  never touches the real source tree. Revised after manager review to drop
  the whole-repo copy, the Git-tracked-file dependency, and the 177-line
  double-`go generate` flow from the original version. Verified it fails
  with a clear message when the checked-in output is made stale (sanity
  checked by corrupting a checksum file and reverting via `git checkout`).
- `just lint` fails repo-wide on this machine due to a pre-existing
  staticcheck/Go-toolchain version mismatch (staticcheck built with go1.26,
  project targets go1.27) — unrelated to this change, not introduced by it,
  and out of scope for #77.
- Final state: `go build ./...`, `go vet ./...`, and `go test ./...` all pass.
  Changes are staged (not committed) in the worktree for review.

## Independent review correction

- Independent review found the test's proof was hollow: it validated the
  `//go:generate` directive text but exercised regeneration via
  `builder.Run(...)` called directly, so replacing the checked-in directive
  with e.g. `//go:generate echo gen-jsonschema` would still pass. Fixed by
  making the test invoke the real directive: `exec.Command("go", "generate",
  "./...")` run twice inside the copied example directory, deleting the
  copied `jsonschema_gen.go`/`jsonschema/*` first so each run must recreate
  them from scratch, then comparing every checked-in artifact byte-for-byte
  after each run.
- To make that possible without a real Git checkout or touching the source
  worktree, the throwaway module now also carries `gen-jsonschema/` (the
  binary the directive invokes via `go run ../../gen-jsonschema/`) and
  `internal/` (what it depends on), copied with `os.CopyFS` alongside the
  root package's non-test `*.go` files, `go.mod`, and `go.sum` — the same
  relative layout as the real repo root, so the example's relative import
  path keeps working unmodified.
- Removed the `internal/builder.Run(...)`/`--no-changes` path entirely (it
  no longer proves anything `go generate` doesn't already cover) and the
  manual `copyDir`/`filepath.WalkDir` tree-copy helper, replaced by
  `os.CopyFS`.
- Reverified: `gofmt -l`, `go build ./...`, and `go test ./...` all pass;
  `go test -run TestAdvertisedExamplesRegenerateCleanly -v` shows both
  subtests passing with two `go generate` invocations each.

## Second review round: two more material fixes

- The artifact list was still discovered by reading the example's
  `jsonschema/` directory at test time (`requireGeneratedArtifacts`
  iterating `os.ReadDir`), so deleting a checked-in file (e.g. a `.sum`)
  just shrank the observed set instead of failing anything — the test
  would silently stop checking the file it should have caught missing.
  Fixed by replacing the dynamic list with an explicit, hardcoded
  `artifacts` slice per example (`jsonschema_gen.go` plus each schema
  `.json`/`.json.sum`) in the `advertisedRegenerableExamples` table, and a
  `requireGeneratedArtifactsExist` helper that `os.Stat`s each expected
  path and fails immediately if any is missing. Verified by deleting
  `examples/enums_stringmode/jsonschema/Paint.json.sum` and confirming the
  test now fails with a clear message, then restoring it via `git
  checkout`.
- `runGoGenerate` shelled out to `go generate ./...` inheriting the test
  process's environment verbatim, so if the parent CI environment sets the
  documented `JSONSCHEMA_NO_CHANGES` variable (any non-empty value; see
  `gen-jsonschema/main.go:111`), every directive would fail immediately
  without writing anything and the test would break for reasons unrelated
  to the code under test. Fixed by setting `cmd.Env = append(os.Environ(),
  "JSONSCHEMA_NO_CHANGES=")` — the empty value overrides any inherited
  non-empty one (Go's `exec.Cmd.Env` keeps the last duplicate key), and
  `os.Getenv` returns `""` for an empty-valued var, which is what
  `main.go`'s `!= ""` check treats as unset.
- Reverified: `gofmt -l`, `go test -run
  TestAdvertisedExamplesRegenerateCleanly -v` both with and without
  `JSONSCHEMA_NO_CHANGES=1` set in the parent shell, and `go test ./...`
  — all pass.

## User correction: drop the live regeneration proof, keep only repo integrity

- User overrode the whole "prove regeneration is honest by executing the
  real directive against a throwaway module copy" design: the test does
  not need to defend against antagonistic or deliberately bogus
  `//go:generate` directives (e.g. swapping in `echo gen-jsonschema`) -
  coding agents are assumed to configure directives honestly. That
  assumption removes the reason for building throwaway-module machinery
  (`os.CopyFS` of `gen-jsonschema/`, `internal/`, root package source,
  `go.mod`/`go.sum`), shelling out to `go generate ./...` twice, or
  reasoning about `JSONSCHEMA_NO_CHANGES` inheritance - none of that was
  needed once the honesty assumption is granted.
- The already-completed live proof (executing the real directive
  end-to-end via the throwaway module, in an earlier iteration of this
  same run, confirmed both a correct first regeneration and a true
  second-run no-op) stands as sufficient evidence that regeneration is
  honest and idempotent; it does not need to be re-run on every `go test
  ./...` invocation going forward.
- Replaced `TestAdvertisedExamplesRegenerateCleanly` with
  `TestAdvertisedExamplesHaveGenerateDirectiveAndArtifacts`, a small
  repository integrity test scoped to exactly the two regressions #77
  fixed: for each of `interfaces_options` and `enums_stringmode`, assert
  (1) at least one real `//go:generate` directive invoking
  `gen-jsonschema` exists among the example's `.go` files, and (2) the
  explicitly named checked-in artifacts exist on disk
  (`jsonschema_gen.go`, the named schema `.json`, and its `.json.sum`).
  No repo copying, no shelling out, no invoking the generator (in-process
  or via `exec`), no Git metadata, no machinery for malformed/bogus
  directives.
- Final proof boundary: this test only guards against the two concrete
  regressions #77 found (a missing entry point, and missing/incomplete
  checked-in artifacts) reappearing later. It does not re-verify
  generator correctness or idempotency each run - that's the generator's
  own test suite's job, and was independently confirmed live during this
  work.
- Reverified after the simplification: `gofmt -l` (clean),
  `go test -run TestAdvertisedExamplesHaveGenerateDirectiveAndArtifacts
  -v .` (both subtests pass), and `go test ./...` (full suite green).
