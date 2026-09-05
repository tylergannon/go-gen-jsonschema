# TypeScript declaration proof

Authority: [definition-of-done.md](definition-of-done.md), the current v1 contract, and the validated Go grammar. Implementation base: main 7991124 plus grammar checkpoint f7cec98. The final implementation is checkpointed only after these checks; no production Node/npm dependency was added to go.mod or go.sum.

## Evidence

- `round-03-go-tests.log`: serialized `go test -p 1 ./...` after the round-02 enum-description correction and full regeneration.
- `round-03-generation.log`: two ordinary `go generate ./...` runs and the environment-driven no-change run produced the same 150-artifact manifest with no generated-file drift.
- `round-03-lint.log`: repository golangci-lint result after regeneration.
- `round-03-typescript-tests.log`: all eleven dedicated conformance claims pass with TypeScript 6.0.3 after the round-02 correction.
- [proof/run-X5rTzB](proof/run-X5rTzB): latest complete fresh consumer, generated TS/schema/Go outputs, direct-backend output, compiler diagnostics, command transcript, and proof metadata. Earlier logs and runs retain pre-correction and failure evidence.
- [projection.md](projection.md): constructor-complete structural projection and induction argument.
- `internal/builder/typegrammar_adapter_test.go`: actual Go source admission, exact constants and source field selection, supported union/presence forms, rejected syntax and unproved mappings.
- `internal/typescript/*_test.go`: constructor, exact-literal, printer/escaping, naming, and invalid-grammar coverage.
- `internal/builder/typescript_output_test.go` and CLI/loader tests: artifact ownership, directory behavior, flag defaults and validation, stale/missing outputs, and target-local loading.

## Independent consensus

- [Round 01](../reviews/20260904-typescript-round-01.md) found a material enum-membership disagreement between schema and TypeScript output.
- [Round 02](../reviews/20260904-typescript-round-02.md) confirmed membership parity and found a material field-local enum-description regression.
- [Round 03](../reviews/20260904-typescript-round-03.md), continuing native reviewer session `9c884e7a-c641-46da-aafa-65207fa8e2f1`, verified both corrections and concluded `only nitpicks remain`.

Consensus accepts the definition of done. The retained nitpicks are a
pre-existing pointer-field `WithEnum` schema-only asymmetry, no direct unit test
for `syntax.ResolveEnum` despite adapter and CLI coverage, and long but correct,
deterministic, injective collision names.

## Observable results

The built CLI was invoked with Node absent from PATH and generated types.ts plus the requested index.ts in a consumer directory separate from its invocation directory. Relative and absolute output paths worked. Reruns were byte-identical. Modified TS failed --no-changes without changing the destination even though schemas were current. Barrel addition/removal/missing-state checks worked, and user-owned files were preserved.

Strict TypeScript compilation of actual CLI output passed independently specified type equalities, positive assignments, expected negative assignments, all three union forms, field-local discriminator identity, Optional/Nullable distinctions, enum modes, refs, time strings, nested composition, and helper-name collision cases. A missing union case produced TS2345, and adding a registered variant caused a previously exhaustive handler to fail. Restoring the original registration restored successful compilation and byte-identical output. The generated Go consumer passed go test ./....

The same consumer independently asserts that emitted JSON Schema enum members match the exact TypeScript obligations: unrelated untyped constants are excluded, typed conversions are included, and computed integer values and escaped strings preserve Go's evaluated values. The scanner, schema renderer and adapter now share resolved enum membership. A Go regression also preserves full-width uint64 JSON literals without narrowing through int or float64.

The round-02 description correction preserves the existing field-local `WithEnum` and `WithStringerEnum` JSON Schema description surface while retaining resolved membership and values shared with TypeScript. Full repository generation left the checked-in examples unchanged.

Fresh direct-backend compiler probes additionally cover empty graphs and barrels, exact versus inexact large integer literals, empty union tags, contextual narrowing of an existing optional discriminator, Unicode naming collisions, unsafe comment terminators and unusual property keys. These probes run in the dedicated npm test/CI lane rather than relying on stored TS snapshots.

This proves structural declaration behavior at the adapter/backend and CLI surfaces. It does not establish runtime bidirectional transport, schema-validation equivalence, or full-width integer transport. Issue #71 owns that separate proof after #57.

## Reproduction

From tests/typescript, run `npm ci --ignore-scripts --no-audit --no-fund` then `npm test`. From the repository root, run `go test -p 1 ./...`, `go generate ./...`, `JSONSCHEMA_NO_CHANGES=1 go generate ./...`, and `golangci-lint run ./...`. Full Go suites must run serially in this worktree because existing builder tests rewrite shared test_run fixtures.
