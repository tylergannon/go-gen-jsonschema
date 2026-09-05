# Adversarial review: native TypeScript generation, round 02

Date: 2026-09-04 (local). Reviewer: Claude Fable 5.1, read-only except this file. Round 01 is retained unchanged at `ephemeral/reviews/20260904-typescript-round-01.md`.

## Review target

The complete current working tree of `codex/typescript-generation` (HEAD f7cec98, implementation base 7991124), including untracked files, re-reviewed against the same authorities as round 01:

- `ephemeral/typescript-generation/definition-of-done.md` (now including the "Review clarification" added after round 01);
- `docs/spec/v1.md`;
- repository instructions (`CLAUDE.md`, `justfile`);
- proof indexed by `ephemeral/typescript-generation/proof-summary.md`, now pointing at `round-02-*.log` and `proof/run-yyMZJF`.

The launch prompt set only operating constraints (artifact path, read-only boundary). No caller narrowing was present or ignored.

## Evidence inspected

Delta since round 01 (by mtime and diff against the round-01 reading):

- New `internal/syntax/enums.go` (`ResolveEnum`, go/types-based membership); `internal/syntax/scan_result.go` now resolves `NewEnumType` registrations through it and drops the `lastTypeName` AST loop; `EnumSet.Values` is now `[]EnumValue{Name, Value constant.Value, Description, Source}`.
- `internal/builder/gen_schema.go`: `discoverEnum` returns `(*EnumSet, error)`; new `renderEnum` shared by `mapEnumType` and the per-field path in `renderStructField`; integer enums render as `PropertyNode[json.Number]` with exact constant text; `nullableSchema` handles the new node. `internal/builder/model.go`: `toJSONValue` emits `json.Number` unquoted.
- `internal/builder/typegrammar.go`: `enum` now copies `EnumSet.Values`; `resolveFieldEnum` falls back to `ResolveEnum`; the duplicated `resolveRegisteredInterfaceField` call from round 01 is gone.
- `internal/builder/typegrammar_adapter_test.go`: added conversion-typed, untyped-decoy, iota, alias, and `uint64` max cases, and a rendered-schema comparison for the `NewEnumType`, `WithEnum`, and `WithStringerEnum` paths.
- `tests/typescript/fixture/types.go`, `consumer.ts`, `check.mjs`: fixture gained `Unrelated`, `Converted`, `Urgent`, `NotPriority`, `Medium`; the lane asserts schema enum arrays and TS literal unions for the same fields (eleven claims).
- `llms.txt`: TypeScript section, both new flags, corrected `-no-changes` description. `ephemeral/typescript-generation/backend-compiler` removed. Worklogs updated with the round-01 correction.
- Unchanged since round 01 and not re-read line by line: `internal/typescript/*`, `typescript_output.go`, `builder.go`, `main.go`, `loader.go`, `grammar.go`, `validate.go`, README, workflow.

Independent verification in this worktree (CLI rebuilt from the current tree into a scratch directory):

| Check | Result |
| --- | --- |
| `go build ./... && go vet ./...` | pass |
| `go test -p 1 ./...` (serial) | pass, all packages ok |
| `golangci-lint run ./...` | 0 issues |
| `go run ./tests/typescript/projection/generate <tmp>` vs committed `projection/generated` | byte-identical |
| `go.mod`/`go.sum` vs HEAD | unchanged |
| Round-01 reproduction (`A Status = "a"; B = "b"; C = Status("c")`) | schema `["a","c"]`, TS `"a" \| "c"`: now agree |
| Extended probe: escaped string, `iota + 1`, `Level(-Low)`, `1 << 40`, `WithStringerEnum` names, `Nullable[Level]`, `Optional[Level]`, cross-package `sub.Kind` with untyped decoy, alias `type Alias = Status` | schema and `types.ts` agree in every case; generated Go compiles |
| `NewEnumType[Level]` for an integer type | rejected by the `~string` constraint at type-check, as before |
| `gen --no-changes` in every `examples/*` directory with the current CLI, then with a CLI built from a `git archive HEAD` copy | see finding 1. Note: `--no-changes` (with or without this work; verified with the HEAD-built CLI in a disposable copy) still rewrites `jsonschema_gen.go` when the Go output differs, so my run without `--validate` rewrote `examples/structs/jsonschema_gen.go`; it was restored to HEAD with `git checkout` and is clean. Pre-existing, not attributable to this branch. |
| `proof/run-yyMZJF/proof.json` | records the current file set, TS 6.0.3, Go 1.27.1, eleven claims; `round-02-typescript-tests.log` shows all passing |

## Findings

### 1. Issue: the shared-enum fix changed JSON Schema output for existing `WithEnum`/`WithStringerEnum` registrations, and the checked-in example outputs are now stale

Classification: issue (unrecorded behaviour change in the pre-existing schema contract, with the repository's own example outputs no longer reproducing).

Evidence. Running the current CLI with `--no-changes` in `examples/stringer_enums` fails with `schema changes detected for types: ApplicationConfig, Task`; a CLI built from HEAD (f7cec98) passes in the same directory. Regenerating in a disposable copy shows the diff: every per-field enum property gained a `description` composed of the field comment plus per-member comments, for example

```text
-"log_level":{"type":"string","enum":["LogDebug","LogInfo","LogWarning","LogError","LogFatal"]},
+"log_level":{"type":"string","description":"LogLevel controls the verbosity of logging\n\nLogDebug: \nLogDebug is for detailed diagnostic information\n\n...","enum":[...]},
-"log_level":{"type":"integer","enum":[0,1,2,3,4]}
+"log_level":{"type":"integer","description":"LogLevel is the minimum log level for this task\n\n0: \nLogDebug is for detailed diagnostic information\n\n...","enum":[0,1,2,3,4]}
```

Cause: `renderStructField` now calls `renderEnum(enumSet, cfg.UseStringer, f.Comments(), f.ID())` (`internal/builder/gen_schema.go` ~1250), and `renderEnum` always builds `enumDescription(description, members, wireValues)` (~617-668). The HEAD per-field path built `PropertyNode` values with no `Desc` at all (HEAD `gen_schema.go` ~1284-1360).

Impact. The DoD clarification asked only that membership and values agree; it did not authorise changing the description surface of the existing JSON Schema for every project using field-local enums. The new output is arguably better (the field comment was previously dropped, and the format mirrors `NewEnumType` output), but it is a wire-visible change to `docs/spec/v1.md`'s "existing generation" that is not mentioned in any worklog, README, or `llms.txt`, is not covered by a builder golden test (no fixture changed, and the suite passed), and leaves `examples/stringer_enums/jsonschema/*.json` and `.sum` files that `go generate` (the documented example workflow in `CLAUDE.md`) will rewrite on the next run. Either regenerate the example outputs and record the description change as intended, or pass an empty description in the per-field path to keep the previous shape. Pre-existing and unrelated: `examples/interfaces_options` (`Owner.json` never checked in) and `examples/enums_stringmode` (`undeclared local type found: TestCase`) fail `--no-changes` identically with the HEAD-built CLI.

### 2. Nitpick: `WithEnum` on a pointer field is silently ignored by the schema renderer but rejected by the adapter

`renderStructField` `continue`s when the render type is not an identifier (`gen_schema.go` ~1227-1229), so `PL *Level` with `WithEnum(Thing{}.PL)` produces `{"type":"integer"}` with no enum and no diagnostic. The adapter instead fails with "enum registration requires a direct named enum type" (`typegrammar.go` ~342-345). With `--typescript` the run fails before any output, so the two artifacts never disagree; without it the registration is dropped silently. Pre-existing schema behaviour, noted because the new adapter diagnostic makes the asymmetry visible.

### 3. Nitpick: `syntax.ResolveEnum` has no direct unit test

The new resolver (`internal/syntax/enums.go`) is exercised only through `internal/builder` adapter tests and the TS lane. A small `internal/syntax` test over a fixture with conversion-typed, untyped, alias-to-named, and alias-to-predeclared cases would pin the membership rules where they are defined.

### 4. Nitpick: collision-suffixed names remain long and opaque

Unchanged from round 01 (`allocateNames`, `internal/typescript/generate.go`). The implementer recorded this as a retained non-material choice; it is still correct and deterministic.

## Round-01 findings status

- Issue 1 (enum membership divergence): resolved. Scanner, renderer, and adapter share `ResolveEnum`; the reproduction and an extended probe agree; the TS lane and adapter tests assert agreement.
- Nitpick 2 (`llms.txt` flags): resolved.
- Nitpick 3 (double `resolveRegisteredInterfaceField`): resolved.
- Nitpick 4 (`backend-compiler` snapshot): resolved.
- Nitpick 5 (collision suffix readability): retained by decision; carried as finding 4.

## Areas re-inspected without findings

- `json.Number` enum nodes: marshalled as bare numbers (`toJSONValue`), handled by `nullableSchema`, and the only `PropertyNode` type switch in the builder covers the new instantiation; the adapter test checks a full-width `uint64` literal.
- Enum string mode: string-kind enums ignore `WithStringerEnum` consistently in both artifacts; integer names mode rejects duplicate underlying values with a diagnostic.
- Alias handling: aliases of a named enum resolve through `types.Unalias`; aliases of predeclared types fail with a source-located diagnostic in both paths.
- Artifact ownership, `--no-changes` for TS, loader `Dir` handling, printer, and union projection: unchanged code, previously verified in round 01, full suite and TS lane still pass.

## Outcome

`material findings remain`
