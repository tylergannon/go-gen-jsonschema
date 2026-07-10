# Worklog: issue 28 planning and fresh-clone repair

Status: complete
Date: 2026-07-10 (America/Costa_Rica)
Workspace: `/Users/tyler/src/go-gen-jsonschema`, branch `main`

## Goal

- Preserve the complete issue #28 Optional/Nullable design and proof plan.
- Prove and route traversal and JSON-property-name prerequisites to GitHub.
- Make a fresh clone or fresh worktree pass its repository tests without
  relying on ignored local generated artifacts.

## Session boundary

The session-worklog skill normally starts substantial edits in a task
worktree. This task already had intentional untracked planning artifacts in the
root checkout before the skill was invoked. Moving at this point would split
one coherent task across worktrees or require staging/copying user-visible
work. Edits therefore remain in the existing workspace; final proof will run
in a separate clean worktree created from the complete candidate tree.

skill_use: session-worklog source=pagerguild/core-tools -> preserve user decisions, bug proof, commands, and clean-worktree evidence for future agents.

## Durable decisions

decision: Ship both `Optional[T any]` and `Nullable[T any]`; ordinary `T`, Optional, and Nullable express required non-null, omittable non-null, and required nullable contracts respectively.
decision: Delete the legacy `jsonschema:"optional"` behavior rather than preserving a compatibility branch.
decision: V1 recognizes either wrapper only as the complete type of a direct named struct field; `Optional[[]T]` is coherent but `[]Optional[T]` is rejected.
decision: Missing `json:",omitzero"` on an Optional field is a generation error because it prevents a real wire-format bug and is cheap to check after classification.
decision: Nullable omission tags are documented and tested but not rejected in V1; `IsZero() == false` neutralizes `omitzero`, and `omitempty` does not omit struct values under `encoding/json` V1.
decision: Nullable generator support is driven by protocol meaning, not every live IR node. V1 supports scalar and inlined-object values, rejects nullable arrays/slices, consts, enums, and explicit refs, and leaves nullable registered interfaces for a separate decision.
decision: Normal nested structs remain inlined. `$ref` is only an explicit legacy-tag/manual-schema path and is not part of the normal Nullable design.

correction: A live `$ref`, const, array, enum, or union IR path does not imply that Nullable must support it.
correction: Cross-model findings remain independently readable; user participates in synthesis before recommendations are folded into the implementation plan.

## GitHub prerequisite reports

- Issue #29: https://github.com/tylergannon/go-gen-jsonschema/issues/29
- Registered-interface failing-test follow-up on #29:
  https://github.com/tylergannon/go-gen-jsonschema/issues/29#issuecomment-4937428864
- Issue #30: https://github.com/tylergannon/go-gen-jsonschema/issues/30

The #29 visibility repair and registered-interface condition repair belong in
one implementation unit: the visibility fix activates the interface failure.
Issue #30 is independent and can be fixed in parallel.

## Proof captured so far

Baseline command:

```text
go test ./...
```

Result: pass in the existing checkout, but only after ignored example
`jsonschema_gen.go` and `jsonschema/` artifacts had been generated locally.
This is not accepted as fresh-clone proof.

Temporary failing tests were added, executed, and removed before returning the
suite to green:

- `TestRegisteredInterfaceIdentifierResolves` failed with
  `undeclared local MarkerInterface type found`; full test and output are in
  the #29 follow-up.
- `TestPropNamesFallsBackToGoNameForEmptyJSONName` expected
  `[]string{"MaxRetries"}` and received `[]string{""}`; full test and output are
  in issue #30.

## Current fresh-clone diagnosis

rule_discovery: examples/.gitignore excludes all generated Go accessors and embedded schema files, so a clean checkout cannot compile examples/structs or examples/providers_rendering.
rule_discovery: internal/syntax/testfixtures/testapp0_simple still invokes the removed `-type` CLI flag, so repository-wide `go generate ./...` fails.
rule_discovery: example `go:generate` directives depend on an externally installed `gen-jsonschema` binary rather than the generator in this checkout.

Planned repair:

1. Track example generated accessors and schema/embed inputs.
2. Make example generation invoke the local generator through `go run`.
3. Remove the obsolete syntax-fixture generation directive.
4. Update example documentation and add a clean-checkout CI proof gate.
5. Run generation, idempotence, tests, and a separate clean-worktree proof.

## Fresh-clone repair implementation

- Deleted `examples/.gitignore` so example `jsonschema_gen.go` accessors and
  `jsonschema/` embed inputs are tracked.
- Replaced every example's dependency on an installed `gen-jsonschema` binary
  with `go run ../../gen-jsonschema/`.
- Removed the obsolete `-type SimpleStruct` generation directive from the
  syntax-only fixture.
- Updated `examples/README.md` to state that generated compile/embed inputs are
  checked in and regenerated from the local checkout.
- Added `.github/workflows/go.yml` to test an untouched checkout, regenerate,
  assert a clean tree, exercise no-change mode, and retest.

Proof command:

```text
go generate ./...
```

Result: pass.

First idempotence attempt:

```text
JSONSCHEMA_NO_CHANGES=1 go generate ./...
```

Result: failed because `internal/builder/messages/assertions.go` hardcoded
`--force`, and the CLI intentionally rejects force plus no-change mode. The
flag dated to an older generator workflow and is unnecessary when generated
artifacts and checksums are current. Removed `--force` from that directive.

Second idempotence attempt with the same command: pass.

### Clean-worktree-only determinism finding

The first synthetic clean-worktree proof passed its initial `go test ./...`
but stopped after regeneration because
`examples/uniontypes/jsonschema_gen.go` changed. The diff only reordered the
generated `Drawing` and `Payment` unmarshalers and their helper functions.

Root cause: `SchemaBuilder.RenderGoCode` iterated `customTypes`, a Go map, when
assembling template input. Existing root-checkout no-change mode could miss
this when the random order happened to match the ignored local artifact.

Repair:

- added `sortedCustomTypeNames` and made `RenderGoCode` consume custom types in
  lexical order;
- added `TestSortedCustomTypeNames`;
- regenerated all artifacts;
- reran ordinary and no-change generation successfully.

This deterministic ordering fix is part of the fresh-clone repair, not issue
#29 or #30.

## GitHub documentation routing

Verified through `gh` that both reports retain the complete failing tests and
observed output:

- #29 registered-interface proof:
  https://github.com/tylergannon/go-gen-jsonschema/issues/29#issuecomment-4937428864
- #30 property-name proof is in the issue body:
  https://github.com/tylergannon/go-gen-jsonschema/issues/30

Added fixing-agent comments that point to the exact plan, independent review,
and worklog sections:

- #29 documentation routing:
  https://github.com/tylergannon/go-gen-jsonschema/issues/29#issuecomment-4937493160
- #30 documentation routing:
  https://github.com/tylergannon/go-gen-jsonschema/issues/30#issuecomment-4937493141

## Final clean-worktree proof

Built a temporary candidate commit from the complete working tree through a
separate Git index. This did not stage or commit changes on `main`. Checked that
candidate out as a detached clean worktree and ran:

```text
git status --porcelain=v1
go test ./...
go generate ./...
git status --porcelain=v1
JSONSCHEMA_NO_CHANGES=1 go generate ./...
go test ./...
git status --porcelain=v1
```

Result:

- initial worktree status: clean;
- initial full tests: pass;
- repository-wide generation: pass;
- status after generation: clean;
- no-change generation: pass;
- final full tests: pass;
- final worktree status: clean.

The temporary proof worktree was removed after success. Candidate proof object:
`f4fc849` (unreferenced temporary commit, not a branch commit).

An unrelated autogenerated `.codex/environments/environment.toml` appeared in
the root workspace during the session. It was preserved untouched and excluded
from the intended candidate tree. The complete clean-worktree sequence above
was repeated with that path absent; it passed and remained clean. Scoped proof
object: `4a701d2` (also an unreferenced temporary commit, not a branch commit).

## Final state

- Fresh checkout/worktree compilation no longer depends on ignored local
  artifacts.
- Generation uses the generator in the checked-out source tree.
- Repository-wide generation and no-change mode are compatible.
- Generated custom-unmarshaler ordering is deterministic.
- CI now enforces clean-checkout tests and clean regeneration.
- Issues #29 and #30 remain open prerequisites; this task documented and routed
  them but did not implement their fixes.

Final existing-workspace closeout after all source changes:

```text
go test ./...        # pass
git diff --check     # pass
```

No example generated accessor or schema/embed input remains ignored.
