# TypeScript generation checkpoint file audit

Base: `f7cec98186bfbe1a06f38380f6656cc503725ffc` (`feat: define authoritative static type grammar`).

## Product and public guidance

- `gen-jsonschema/main.go`, `internal/builder/builder.go`, and `internal/builder/typescript_output.go`: CLI flags, builder integration, preflight, ownership, drift, and output mutation.
- `internal/builder/typegrammar.go`, `internal/syntax/enums.go`, `internal/syntax/loader.go`, `internal/syntax/scan_result.go`, `internal/builder/gen_schema.go`, `internal/builder/model.go`, and `internal/typegrammar/grammar.go`: checked source lowering, shared exact enum values, target-local loading, and exact schema literals.
- `internal/typescript/`: native TypeScript AST, lowering, naming, and printer.
- `README.md` and `llms.txt`: flags, artifact layout, imports, ownership, no-change behavior, and structural-only limits.

## Tests and CI

- `gen-jsonschema/main_test.go`, `internal/builder/typegrammar_adapter_test.go`, `internal/builder/typescript_output_test.go`, `internal/syntax/loader_test.go`, and `internal/typescript/*_test.go`.
- `tests/typescript/`: pinned TypeScript 6.0.3 conformance lane, independent source and backend projections, strict consumer obligations, and reproducible fixture inputs.
- `.github/workflows/typescript.yml`: dedicated TypeScript compiler lane.

## Retained operating and proof artifacts

- `ephemeral/typescript-generation/`: accepted definition of done, projection argument, round-03 proof logs, latest `proof/run-X5rTzB`, and earlier attempts retained for provenance.
- `ephemeral/reviews/20260904-typescript-*`: exact minimal review prompts, three review rounds, and native session metadata.
- `ephemeral/worklog/20260904-ts-{adapter,backend,cli}.md` and `ephemeral/worklog/20260904-typescript-generation.md`.

## Exclusions and cleanliness

- `go.mod` and `go.sum` are unchanged; production has no Node/npm dependency.
- Checked-in examples, builder fixtures, and generated schema/Go outputs have no diff after two ordinary generations and the environment no-change run.
- `tests/typescript/node_modules/` and `tests/typescript/projection/generated/` are ignored local build inputs/outputs and are excluded from the checkpoint.
- No generated CLI binary, Go test binary, shared library, or dependency directory is in the candidate set.
- The original research worktree, inspiration material, and root-checkout files are outside this worktree and outside the checkpoint.
