# Adversarial review: native TypeScript generation, round 01

Date: 2026-09-04 (local). Reviewer: Claude Fable 5.1, read-only except this file.

## Review target

The complete current working tree of `codex/typescript-generation` (HEAD f7cec98, implementation base 7991124), including untracked files, reviewed against:

- `ephemeral/typescript-generation/definition-of-done.md` (authoritative completion criteria);
- `docs/spec/v1.md` (accepted v1 Go/JSON contract);
- repository instructions (`CLAUDE.md`, `justfile`);
- proof indexed by `ephemeral/typescript-generation/proof-summary.md`.

The launch prompt set only operating constraints (artifact path, read-only boundary, proof index). It did not narrow subject matter, predict findings, or request a verdict, so no caller narrowing was ignored.

## Evidence inspected

Source, all read in full:

- `internal/builder/typegrammar.go` (adapter, `TypeDefinitions`), `internal/builder/typescript_output.go` (artifact plan/apply), `internal/builder/builder.go` diff (`Run` integration).
- `internal/typescript/ast.go`, `generate.go`, `printer.go` (backend).
- `internal/typegrammar/grammar.go` (doc-only change vs HEAD) and `validate.go`.
- `internal/syntax/loader.go` diff (directory-aware `Load`), `gen-jsonschema/main.go` diff, `README.md` diff, `.github/workflows/typescript.yml`.
- Surrounding builder code: `SchemaMethods`, `mapType`, `mapEnumType`, `discoverEnum`, `resolveRegisteredInterfaceField`, EnumV1/TypeProvidersMap keying, `syntax.StructField` helpers (`Fields`, `Skip`, `Wrapper`, `HasJSONOption`, `JSONTag`), `scan_result.go` constant collection.
- Tests: `typegrammar_adapter_test.go`, `typescript_output_test.go`, `generate_test.go`, `printer_test.go`, `loader_test.go`, `main_test.go`; the TS lane `tests/typescript/check.mjs`, `consumer.ts`, `missing-case.ts`, `fixture/*.go`, `projection/generate/main.go`, `projection/consumer.ts`, tsconfigs, `package.json`.
- Proof: `proof-summary.md`, `projection.md`, `final-go-tests.log`, `final-lint.log`, `final-typescript-tests.log`, `proof/run-XYiyEI/{proof.json,commands.log,consumer/generated/types.ts}`, the three other retained runs' `proof.json`, and the four worklogs.

Independent verification performed in this worktree:

| Check | Result |
| --- | --- |
| `go build ./... && go vet ./...` | pass |
| `go test -p 1 ./...` (serial, as the DoD requires) | pass, 13 packages ok, no failures |
| `golangci-lint run ./...` | 0 issues |
| `staticcheck ./...` | could not run: installed staticcheck cannot read Go 1.27 export data (toolchain mismatch, not a code defect) |
| `go run ./tests/typescript/projection/generate <tmp>` vs committed `tests/typescript/projection/generated` | byte-identical |
| `go.mod`/`go.sum` vs 7991124 | unchanged; no Node/npm/JS dependency added |
| `node_modules`, `projection/generated`, proof dirs | git-ignored as documented |
| Fresh consumer module built against this worktree with the built CLI, `--typescript` relative dir, cross-package dependency type | generated `types.ts` and schemas; dependency `Detail` emitted with description |
| Grep for `Chdir` in production code, and for Node invocation in `*_test.go` | none |

The retained run `run-XYiyEI` records this exact working-tree file set, TypeScript 6.0.3, Go 1.27.1, and all ten claims passing with Node removed from the generator's PATH. The projection argument in `projection.md` matches the implemented cases in `generate.go` and the sealed constructor sets in `validate.go`.

## Findings

Findings are limited to the five most severe; everything below was inspected but not all of it produced a finding.

### 1. Issue: the same run emits contradictory enum member sets in the JSON Schema and in `types.ts`

Classification: issue (incorrect implementation of a shared requirement; observable contradiction between two artifacts of one generation).

Evidence. The adapter enumerates enum members from `go/types` package-scope constants of the exact named type (`internal/builder/typegrammar.go`, `enum`, lines ~470-500). The existing scanner and renderer instead use an AST heuristic that (a) inherits the previous explicit type across any later untyped constant in the same `const` block (`internal/syntax/scan_result.go` lines ~452-478, `lastTypeName`), and (b) only recognises constants whose ValueSpec has an explicit type identifier, so conversion-typed constants are missed (`internal/builder/gen_schema.go`, `discoverEnum`, lines ~489-501).

Reproduction (run from a fresh module that `replace`s this repository, using the CLI built from this worktree):

```go
type Status string

const (
	A Status = "a"
	B        = "b" // untyped string constant, not a Status
)

const C = Status("c") // typed Status by conversion

type Thing struct {
	S Status `json:"s"`
}
```

with `NewJSONSchemaMethod(Thing.Schema)` and `NewEnumType[Status]()`, then `gen --typescript ts --pretty`:

```text
jsonschema/Thing.json:  "enum": ["a", "b"]
ts/types.ts:            export type Status = "a" | "c";
```

Impact. The TypeScript projection follows `docs/spec/v1.md` ("Numeric mode uses underlying constant values"; enums are the constants of the Go type) and is the correct set. The schema is the pre-existing renderer's set and is wrong in both directions for this input. But the DoD's promise is `types.ts` generated "alongside ordinary generation", and a consumer now receives two generator-owned contracts that admit different wire values with no diagnostic. This gap is created by the implementation choice to derive members from a different source than the renderer without reconciling or diagnosing the difference; the DoD item 2 requirement to "retain exact constants" is met for TS alone, but item 1's "alongside ordinary generation" is not coherent. Either the scanner/renderer must adopt the `go/types` member set (the correct fix, and within the repository's v1 obligations), or the adapter must fail with a diagnostic when its member set differs from the set the schema will publish. No test in either layer asserts the two sets agree.

### 2. Nitpick: `llms.txt` CLI reference does not mention the new flags

`docs/spec/v1.md` names `llms.txt` as the current user guidance. Its "CLI reference" block (lines ~461-468) lists `-pretty`, `-target`, `-no-changes`, `-force`, `--validate`, `--formats`, and its `-no-changes` description still says "if schema JSON would change". `README.md` was updated (flags, import example, type-only limits, #71 pointer), which satisfies DoD item 7, but the two public references now disagree about `--no-changes` semantics and about which flags exist.

### 3. Nitpick: redundant double resolution in the enum branch of `fieldValue`

`internal/builder/typegrammar.go`, enum branch (~lines 300-306): `resolveRegisteredInterfaceField(owner, field)` is called twice, once for the error and once for the value. A single call with both results is equivalent and avoids re-running interface validation. No behavioural impact.

### 4. Nitpick: retained `backend-compiler` snapshot duplicates the projection lane

`ephemeral/typescript-generation/backend-compiler/` is a stored TS snapshot from before the projection lane was promoted into `tests/typescript/projection`. `proof-summary.md` explicitly says those probes now run fresh in the npm lane rather than relying on stored snapshots, so the directory is unindexed residue that can drift from the generator. Harmless, but it is not part of the indexed proof.

### 5. Nitpick: collision-suffixed names are long and opaque

`allocateNames` (`internal/typescript/generate.go`) resolves a base-name collision by appending a hex encoding of `packagePath + NUL + name` to every colliding definition (for example `_u96EA_$6578616d706c652e636f6d...`). This is deterministic and injective as the DoD requires, so it is correct; but the identifiers are unreadable at any real package path length. A shorter deterministic scheme (for example a stable short hash, or the last package path segment with a numeric suffix on further collision) would meet the same requirement with usable names.

## Areas inspected without material findings

- Artifact ownership and no-change behaviour: preflight before any output mutation, header-based ownership on both read and apply, atomic temp-file rename, owned-barrel removal only, `--no-changes` never touching the destination, relative paths resolved against the invocation CWD with no `Chdir` anywhere. Matches DoD item 5.
- Union projection `Omit<Impl, disc> & { disc: tag }`: preserves shared payload declarations, preserves singleton tags, supports `I`, `Optional[I]`, `[]I`; the validator rejects incompatible existing discriminator properties rather than silently overwriting them, consistent with the spec's runtime rejection of conflicting discriminators.
- Exact literals: `constant.Float64Val(constant.ToFloat(v))` exactness check rejects 2^53+1 and prints the exact decimal; string-mode enums use constant names and never `String()` (confirmed by the fixture's decoy `String()` and the negative consumer assertion).
- Printer: precedence-aware parenthesisation, all keys JSON-quoted, comments sanitised for `*/`, control characters and U+2028/2029, deterministic ordering, `export {};` for empty modules.
- Diagnostics: providers, explicit schema refs, `json:",string"`, custom `MarshalJSON`/`MarshalText`/`Unmarshal*` methods, maps/chans/funcs/inline interfaces, wrappers in non-direct positions, and missing `omitzero` on `Optional` all fail with source or field context; no `any`/`unknown` fallback exists in the backend.
- Loader change: `syntax.Load` now sets `packages.Config.Dir` for directory targets, keeps import-path loading for non-existent paths, and is regression-tested; full suite passes.
- TS lane: pinned TypeScript 6.0.3 via lockfile, `strict`, `exactOptionalPropertyTypes`, `noUncheckedIndexedAccess`; Node removed from the generator's PATH; missing-case and added-variant exhaustiveness proofs use real compiler diagnostics; CI workflow uploads the retained proof.

## Outcome

`material findings remain`
