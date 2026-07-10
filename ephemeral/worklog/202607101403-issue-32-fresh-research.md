# Worklog: issue 32 fresh research and implementation

Status: research in progress
Date: 2026-07-10 (America/Costa_Rica)
Workspace: `/Users/tyler/src/go-gen-jsonschema-issue-32`
Branch: `codex/issue-32-optional-nullable`
Base: `origin/main` at `efab9952c5ea71fc7ff291e71fb13f3fd4bfafd0`

## Goal

- Re-research issue #32 rather than treating its existing decisions or proof
  section as established fact.
- Define implementation checkpoints with independent third-party review between
  risky design or implementation stages.
- Build behavioral proof through the real generator, generated code, schema
  validator, and JSON runtime path; unit tests and CI are closeout hygiene.

## User constraints and corrections

correction: Pull the latest `origin/main` before creating the issue worktree.
correction: The issue's existing proof-of-work description is not accepted as
sufficient; replace it with concrete, inspectable runtime evidence.
correction: The enabled OpenAI documentation/plugin surface must not be treated
as uninstalled merely because its MCP tools are absent from one task's callable
registry. An unnecessary global MCP entry was added, immediately removed, and
no configuration change remains.
rule_discovery: Treat issue #32 as an untrusted proposal and verify language,
schema, OpenAI, generator, and decoder assumptions from primary evidence.

## Session start

- Root baseline before other work: `go test ./...` passed.
- `git fetch --prune origin main` confirmed remote head
  `efab9952c5ea71fc7ff291e71fb13f3fd4bfafd0` (`fix: repair exported field
  traversal (#35)`).
- Created this worktree directly from that remote head; the unrelated untracked
  root `.codex/` directory remains untouched.
- Issue #32 is open at <https://github.com/tylergannon/go-gen-jsonschema/issues/32>.

skill_use: proof-of-work source=pagerguild/core-tools -> define proof around the
running generator and generated consumer behavior, with uploaded artifacts at
PR time.
skill_use: session-worklog source=pagerguild/core-tools -> retain research,
corrections, review gates, proof results, and delivery state.

## Initial concern

The issue body mixes user semantics, public API decisions, implementation
prescriptions, exclusions, documentation scope, and a large checklist. Research
must determine which claims are necessary, which are merely plausible, and
which make the first change too broad to review or prove honestly.

## Review checkpoints

Provisional until research finishes:

1. Contract checkpoint: independent review of the verified wire semantics,
   Go runtime behavior, compatibility/breaking-change policy, and V1 scope.
2. Architecture checkpoint: independent review of wrapper recognition,
   requiredness/nullability IR boundaries, and generated-decoder strategy,
   backed by focused red tests or a spike.
3. Runtime checkpoint: independent review of the implementation diff plus real
   generator/generated-consumer evidence before broad fixtures and docs settle.
4. Release checkpoint: independent review of the exact PR head, uploaded proof,
   compatibility notes, generation idempotence, and green CI before merge.

## Evidence log

- Baseline `go test ./...`: pass.
- Fresh `origin/main` and worktree base: confirmed as above.
- Current main contains the issue #29 and #30 prerequisite repairs.
- Official JSON Schema documentation confirms that required-key presence and
  acceptance of JSON null are separate constraints.
- Current OpenAI Structured Outputs documentation confirms that every property
  must be required and documents a union with null as its optional-value idiom.
- Go 1.26.5 local sources and package documentation confirm that `,omitzero`
  consults `IsZero() bool`.
- Added `docs/design/issue-32-checkpoint-plan.md` with verified claims,
  unaccepted issue-body prescriptions, four independent-review gates, and an
  executable generator-to-generated-consumer proof contract.
- User correction: keep the work grounded in the concrete public
  `Optional[T]` / `Nullable[T]` feature and calibrate ceremony to a medium-sized
  generator change. The earlier standalone legacy-behavior example was removed
  because it did not prove the feature being built.
- Added `docs/design/issue-32-definition-of-done.md` as the controlling concise
  acceptance contract. Requested an independent Claude Fable review focused on
  direction, scope, and whether the proof demonstrates real public behavior.
- Claude Fable round 1 found no critical or bug issues and two design
  ambiguities: the live OpenAI response needed to pass generated validation and
  decode, and leftover legacy optional tags needed explicitly inert semantics.
  Both were incorporated along with four closeout clarifications.
- Claude Fable round 2 reached consensus on direction and proof with no critical,
  bug, or design findings. Its only nit, applying integer-width tests to both
  ordinary and wrapped fields, was incorporated.
- Converted the accepted definition of done into the controlling actionable
  implementation artifact with ordered file-level work, checkboxes, exit
  commands, concrete proof artifacts, and two third-party review gates.
- User correction: a live Structured Outputs call is unnecessary ceremony and
  must not require an API key. Removed that gate from the controlling artifact.
  OpenAI compatibility guidance now follows the published rule directly: all
  properties are required, and optional-value semantics use a union with null.

doc_lookup: https://developers.openai.com/api/docs/guides/structured-outputs -> strict Structured Outputs requires all fields and recommends a null union for optional-value semantics; no live probe required.

## Implementation and local proof

- Added public `Optional[T]` and `Nullable[T]` runtime wrappers with presence,
  zero/empty preservation, null rejection/acceptance, and transactional decode.
- Added canonical direct-field wrapper classification in `internal/syntax`,
  including import aliases and actionable rejection of nested/aliased placement.
- Added Optional requiredness, Nullable scalar/object schema rendering,
  Optional `omitzero` enforcement, and transactional generated interface decode.
- Added `examples/optionality` as the executable public consumer proof. Its
  deterministic transcript covers schema shape, validation, wrapper state,
  round trips, interfaces, mutation safety, and real-generator negative cases.
- Removed live `jsonschema:"optional"` behavior and migrated fixtures/guidance.
- Added ordinary/Optional/Nullable numeric coverage for the generator's
  supported integer widths, named scalars, float32, and float64.

proof: `go generate ./...` -> pass.
proof: `go test ./...` -> pass after repository-wide generation.
proof: `go run ./examples/optionality/cmd/proof` -> exact match with committed `proof/expected.json`.

## Finished-product review: Fable

Review: `ephemeral/reviews/2026071016-issue-32-finished-fable.md`

Accepted findings:

- bug: Nullable legacy enums produced a null type union constrained by a
  non-null enum. `nullableSchema` now rejects enum/const property nodes, with a
  real negative generator fixture.
- bug: slice `byte`/`uint8` used array schemas although encoding/json uses
  base64 strings. Slice rendering now emits a string schema; the executable
  proof decodes and re-marshals an Optional byte slice.
- design: the legacy registered-interface path lacked the explicit Nullable
  guard. It now returns the same actionable unsupported-interface diagnostic.
- design: public guidance claimed unproved Optional enum/provider support.
  Guidance now matches the accepted V1 scope.
- design/nit: negative placement proof was mislabeled and incomplete. It now
  separately covers nested, container, alias, defined, embedded, and root
  wrapper placements, plus Nullable enum/interface rejection.
- nit: added the legacy-tag migration warning to README.

Deferred until delivery: the review correctly noted the branch had no commit,
so clean-tree/fresh-checkout proof was not yet meaningful. Commit and exact-head
closeout remain the next delivery stage.

## Follow-up review: Sonnet

Review: `ephemeral/reviews/2026071017-issue-32-followup-sonnet.md`

- Sonnet independently verified every Fable fix, the full suite, proof
  transcript, idempotent generation, and no-change generation.
- Accepted bug: Nullable byte/uint8 slices bypassed slice rejection after their
  ordinary schema was correctly collapsed to a base64 string. Nullable now
  rejects array/slice Go shapes before schema transformation, and a dedicated
  real-generator negative fixture proves it.
- Accepted proof nit: added committed negative fixtures for Nullable explicit
  refs and providers, in addition to the existing enum/interface/slice cases.
- Deferred nit: the duplicated two-line Nullable-interface guard is clearer in
  place than a helper that would obscure the two distinct registration paths.
- Cleanup: removed an untracked root `jsonschema_gen.go` left by Sonnet's
  throwaway probe after the review process exited.

correction: User rejected byte-array handling as irrelevant scope expansion for
an LLM structured-output generator. Removed the added byte/rune/uintptr surface,
the Optional byte-slice behavior, and its proof fixture. Issue 32 remains about
presence and nullability for meaningful structured-output fields.

Final pre-delivery Sonnet review:
`ephemeral/reviews/2026071018-issue-32-final-sonnet.md`. It found no remaining
critical or bug findings. Its one design finding was a README example that
listed `phone` in `required` without showing the property; the example now
includes the actual `type: ["string", "null"]` property schema.

## Committed closeout

- Candidate commit before final evidence update:
  `acf48c54aa72dd62810b66c0f527c8956f1ac9c5`.
- In the clean task worktree: `go generate ./...`, `git diff --exit-code`,
  `JSONSCHEMA_NO_CHANGES=1 go generate ./...`, `go test ./...`, and
  `git diff --check` all passed; the tree remained clean.
- Fresh detached worktree at the same commit: repository-wide generation,
  no-change generation, full tests, and
  `go run ./examples/optionality/cmd/proof` all passed; proof output matched
  `examples/optionality/proof/expected.json` byte-for-byte and the fresh tree
  remained clean. The temporary worktree was removed afterward.

doc_lookup: https://json-schema.org/understanding-json-schema/reference/object -> required presence and null are distinct.
doc_lookup: https://json-schema.org/understanding-json-schema/reference/type -> type arrays express JSON type unions.
doc_lookup: https://developers.openai.com/api/docs/guides/structured-outputs -> all fields required, null unions, root and anyOf restrictions.
doc_lookup: https://pkg.go.dev/encoding/json -> omitzero uses IsZero.

skill_use: claude-codex-consensus source=pagerguild/core-tools -> independent
review at contract, architecture, implementation, and exact-head proof gates.
