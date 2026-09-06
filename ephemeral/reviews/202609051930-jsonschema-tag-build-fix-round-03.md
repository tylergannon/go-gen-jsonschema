# Adversarial review: jsonschema-tag build fixes, round 03

## Review target

The entire current branch `claude/gallant-austin-9fd4e3` at `f9c866b`,
including the original example fixes, the free-function-root generator and
proof changes, and the round-02 response. I reviewed it against the same
authoritative goal and contracts used in the first two rounds: repair the three
reported `jsonschema`-tag compile errors without deleting supported pointer-root
behavior, add durable regression proof, and resolve the material findings from
the preceding review rather than silently discarding a registered root from a
requested output.

## Evidence inspected

- Repository `AGENTS.md`, including the registration contract for free-function
  roots and the `--validate` output contract; `Declare`'s public contract in
  `declare.go`
- Full `origin/main...HEAD` diff and all four branch commits; all affected
  examples, generated artifacts, builder root classification and templates,
  TypeScript lowering, fixture harness, CI workflow, and `justfile`
- `ephemeral/worklog/202609051649-fix-jsonschema-tag-build-errors.md` and review
  rounds 01 and 02
- The current GitHub follow-ups #90, #91, and #92, fetched directly with
  `gh issue view`
- Baseline `go test ./...`: exit 0
- `just build-tagged`, `go build ./...`, `go vet ./...`,
  `go run ./internal/cmd/doc-gen -check`, and `git diff --check
  origin/main...HEAD`: exit 0
- Independent disposable-fixture reproduction using the current CLI:
  `go run ./polytype gen -target <entrypoints-copy> -validate -force` emitted
  `PointerFuncTypeSchema` but no `PointerFuncType` validator or compiled-schema
  variable; after `go mod tidy`, that fixture's `go test ./...` still exited 0
  because its three ordinary roots exercise the shared validation machinery

The tagged-build regression itself is fixed: the three public example packages
compile under the generation tag, the named-pointer examples retain their
standalone schemas and callable accessors through free-function registration,
and the local/CI `build-tagged` checks cover the reported example-package class.
The TypeScript half of round 02's finding is also fixed and unit-tested.

## Findings

### 1. issue — Round 02's validation finding remains unresolved; the generator still succeeds while omitting a requested root

The new root split puts method-capable entries in `SchemaMethods()` and named
pointer/interface free-function entries in `SchemaFreeFuncs()`
(`internal/builder/gen_schema.go:727-760`). TypeScript lowering now visits both
sets (`internal/builder/typegrammar.go:31-42`), but validation still does not:
`HasNonRenderedTypes()` examines only `SchemaMethods()`
(`internal/builder/gen_schema.go:661-669`), and the compiled-schema and
`ValidateJSON` template blocks range only over `.SchemaMethods`
(`internal/builder/schemas.go.tmpl:61-98,145-176`). The new runtime proof calls
only `PointerFuncTypeSchema` and never requests or asserts validation
(`internal/builder/testfixtures/entrypoints/entrypoints_test.go:8-27`).

This is not hypothetical. Regenerating a disposable copy of that exact fixture
with `-validate -force` produced a `PointerFuncTypeSchema` function at generated
line 93 but no validator or compiled-schema variable for `PointerFuncType`; the
command nevertheless succeeded. That contradicts the repository contract that
`--validate` generates validation for non-rendered registered schemas
(`AGENTS.md:98-100`) and the public contract that `Declare` accepts a
free-function root and the same chained options as a method root
(`declare.go:12-17`). Go cannot attach a method to this named pointer base, but
that requires an explicit free-function validation API or an actionable
generation-time rejection, exactly as round 02 specified; silent success and
omission is still the material failure.

Filing #92 records the debt but does not make the current implementation or
proof complete. The branch should either implement one of those explicit
contracts and add a real generated-fixture test for it, or reject `--validate`
with this root class before mutating outputs.

### 2. issue — Follow-up #90 publishes a false test-coverage premise that the corrected worklog already disproves

The live issue says `testfixtures/indirecttypes` is not in `TestBasic`, calls it
orphaned, and asks to add it back. In the current authoritative harness it is
already a case at `internal/builder/basic_test.go:113-134`; that case generates
the fixture, checks all listed JSON goldens, then runs real `go build ./...` and
`go test ./...` through the shared harness at lines 49-99. The worklog itself
correctly retracts the earlier claim and explains that the accessor omission is
hidden by `SchemaMethods()` filtering, not by absence from the harness
(`ephemeral/worklog/202609051649-fix-jsonschema-tag-build-errors.md:190-205`).
Yet the published issue body was not corrected and still asserts the opposite.

Impact: the branch's claimed follow-up/proof state directs future work from a
known-false premise and can lead a maintainer to delete an active fixture or
spend time "adding back" a case that already runs. Update #90 to state the
actual gap: the fixture is active and its JSON is checked, but the generated Go
accessor set is not asserted, so invalid method-root registrations are silently
filtered before the subsequent build/test steps.

## Outcome

material findings remain
