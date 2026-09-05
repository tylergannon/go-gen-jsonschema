# Adversarial review: native TypeScript generation, round 03

Date: 2026-09-04 (local). Reviewer: Claude Fable 5.1, read-only except this file. Rounds 01 and 02 are retained unchanged.

## Review target

The complete current working tree of `codex/typescript-generation` (HEAD f7cec98, implementation base 7991124), including untracked files, re-reviewed against the same authorities as rounds 01 and 02:

- `ephemeral/typescript-generation/definition-of-done.md` (including the post-round-01 "Review clarification");
- `docs/spec/v1.md`;
- repository instructions (`CLAUDE.md`, `justfile`);
- proof indexed by `ephemeral/typescript-generation/proof-summary.md`, now pointing at `round-03-*.log` and `proof/run-X5rTzB`.

The launch prompt set only operating constraints (artifact path, read-only boundary). No caller narrowing was present or ignored.

## Evidence inspected

Delta since round 02 (files newer than the round-02 artifact, excluding regenerated-identical fixtures):

- `internal/builder/gen_schema.go`: `renderEnum` gained a `withDescriptions` parameter; `enumDescription` returns `""` when disabled; the field-local path in `renderStructField` now calls `renderEnum(enumSet, cfg.UseStringer, "", false, f.ID())` (~1253) while `mapEnumType` keeps descriptions (~609).
- `internal/builder/typegrammar_adapter_test.go`: the rendered-schema assertion now requires empty descriptions for `WithEnum`/`WithStringerEnum` properties and description-bearing output for the globally registered enum (~203-207), with fixture comments placed on the field-local fields to make the distinction observable.
- Every `examples/*` and `internal/builder/test_run/*` generated artifact was regenerated (mtimes 18:32-18:34); `git status` shows none of them modified, so regeneration was byte-identical to the checked-in tree.
- Proof: `proof-summary.md`, `round-03-generation.log` (two `go generate ./...` runs and one `JSONSCHEMA_NO_CHANGES=1` run, same 150-artifact manifest), `round-03-go-tests.log`, `round-03-lint.log`, `round-03-typescript-tests.log` (eleven claims), `proof/run-X5rTzB/{proof.json,commands.log}`; the three updated worklogs.
- Unchanged since round 02 and not re-read line by line: `internal/typescript/*`, `internal/builder/typegrammar.go`, `typescript_output.go`, `builder.go`, `main.go`, `loader.go`, `enums.go`, `scan_result.go`, `grammar.go`, `validate.go`, README, `llms.txt`, workflow, TS lane sources.

Independent verification in this worktree (CLI rebuilt from the current tree into a scratch directory; all example runs performed in a disposable rsync copy so the worktree was not touched):

| Check | Result |
| --- | --- |
| `go build ./... && go vet ./...` | pass |
| `go test -p 1 ./...` (serial) | pass, all packages ok |
| `golangci-lint run ./...` | 0 issues |
| `gen --no-changes` in each `examples/*` (with each directory's own `go:generate` flags) | all pass, including `stringer_enums`, which failed in round 02. `interfaces_options` (`Owner.json` never checked in) and `enums_stringmode` (`undeclared local type found: TestCase`) still fail identically with a HEAD-built CLI; pre-existing, unrelated |
| `go run ./tests/typescript/projection/generate <tmp>` vs committed `projection/generated` | byte-identical |
| Enum-parity probe from round 02 (typed conversion, untyped decoy, `iota + 1`, negative, `Nullable[Level]`, `Optional[Level]`, cross-package `sub.Kind`) | schema and `types.ts` still agree; generated Go compiles |
| Tampered `types.ts` then `gen --typescript ts --no-changes` | fails naming `ts/types.ts`; destination unmodified |
| `gen --typescript-barrel` without `--typescript` | fails with the documented diagnostic, exit 1 |
| `go.mod`/`go.sum` vs HEAD | unchanged |

## Findings

No material findings remain. The nitpicks below are the only items retained.

### 1. Nitpick: `WithEnum` on a pointer field is silently ignored by the schema renderer but rejected by the adapter

Unchanged from round 02. `renderStructField` `continue`s when the render type is not an identifier (`internal/builder/gen_schema.go` ~1227-1229), so `PL *Level` with `WithEnum(Thing{}.PL)` renders `{"type":"integer"}` with no enum and no diagnostic in a schema-only run; with `--typescript` the adapter rejects it first (`internal/builder/typegrammar.go` ~342-345). Pre-existing schema behaviour; the artifacts never disagree because the TS run fails before output.

### 2. Nitpick: `syntax.ResolveEnum` has no direct unit test

`internal/syntax/enums.go` is exercised only through `internal/builder` adapter tests and the TS lane. A small `internal/syntax` test over conversion-typed, untyped, alias-to-named, and alias-to-predeclared cases would pin the membership rules where they are defined.

### 3. Nitpick: collision-suffixed names remain long and opaque

Unchanged from rounds 01 and 02 (`allocateNames`, `internal/typescript/generate.go`); recorded by the implementer as a retained non-material choice. Still correct, deterministic, and injective.

## Prior-round findings status

- Round 01 issue 1 (enum membership divergence): resolved in round 02, still holds.
- Round 02 issue 1 (field-local enum descriptions changed, stale example outputs): resolved. The field-local path suppresses descriptions, the adapter test pins the distinction from global enum output, and full regeneration leaves every checked-in example and fixture byte-identical.
- Round 01 nitpicks 2-4: resolved in round 02. Round 01 nitpick 5 and round 02 nitpicks 2-4: carried above.

## Areas re-inspected without findings

- Global `NewEnumType` descriptions: `examples/enums` output unchanged; `mapEnumType` still passes type and member comments.
- `json.Number` enum nodes, string-mode duplicate rejection, alias handling: unchanged code, adapter test and probe still pass.
- Artifact ownership, TS `--no-changes`, barrel validation, loader `Dir` handling, printer, and union projection: unchanged since round 01; suite, lane, and direct probes pass.

## Outcome

`only nitpicks remain`
