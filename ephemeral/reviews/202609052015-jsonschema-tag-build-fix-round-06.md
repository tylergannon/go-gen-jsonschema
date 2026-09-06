# Adversarial review: jsonschema-tag build fixes, round 06

## Review target

The entire current branch `claude/gallant-austin-9fd4e3` at `52658c4`,
including the original three example repairs, the tagged-build gate, the
free-function-root generation and TypeScript changes, and all responses to
review rounds 01 through 05. I reviewed it against the same authoritative
contracts used in those rounds: repair the reported generation-tag failures
without discarding supported registrations, preserve the documented legacy and
fluent registration APIs, fail explicitly instead of emitting partial or
uncompilable output, and retain behavioral proof at the claimed acceptance
surface.

## Evidence inspected

- Repository `AGENTS.md`; the public `Declare`, `NewJSONSchemaFunc`,
  `NewJSONSchemaBuilder`, and `NewInterfaceImpl` contracts; README registration
  and TypeScript guidance; the Go receiver-base rule in the installed Go
  specification (`$GOROOT/doc/go_spec.html:2995-3002`)
- Full `origin/main...HEAD` diff and all seven branch commits; scanner marker
  parsing and type resolution, builder root classification and templates,
  TypeScript lowering, affected examples and generated artifacts, fixture
  harness, CI workflow, and `justfile`
- `ephemeral/worklog/202609051649-fix-jsonschema-tag-build-errors.md` and review
  rounds 01 through 05
- Current GitHub issues #90, #91, and #92 fetched with `gh issue view`
- Fresh `go test ./...`, `just build-tagged`, `go build ./...`, `go vet ./...`,
  `go run ./internal/cmd/doc-gen -check`, and
  `git diff --check origin/main...HEAD`: all exited 0
- Direct compiler reproduction for a forwarding pointer definition:
  `type P *int; type Q P; func (Q) M() {}` fails with
  `invalid receiver type Q (pointer or interface type)`

Round 05's two findings are fixed at their named acceptance surfaces.
`NewJSONSchemaBuilder` registrations are now distinguished from one-argument
free-function roots and rejected for the directly declared invalid receiver
shape, while the retained entrypoints fixture now compiles, calls, and inspects
the generated interface-root accessor. The interface `--validate` rejection is
also covered. The original three public example packages continue to compile
under the generation tag.

## Findings

### 1. issue — The receiver classifier checks declaration syntax, not the underlying type, so forwarding pointer/interface definitions still generate illegal methods

`hasInvalidMethodReceiverBase` claims to test the named type's *underlying*
type, but for ordinary named types it switches directly on
`ts.Type().Expr()` and recognizes only an immediate `*dst.StarExpr` or
`*dst.InterfaceType` (`internal/builder/gen_schema.go:714-731`). `TypeSpec.Type`
returns the declaration's original AST expression without resolving named
definitions (`internal/syntax/node_wrappers.go:825-827`). Therefore a valid
registration source such as:

```go
type P *int
type Q P

func QSchema(Q) json.RawMessage { panic("not implemented") }
var _ = polytype.Declare(QSchema)
```

is misclassified. `Q`'s declaration expression is the identifier `P`, so the
helper returns false even though Go resolves `Q`'s underlying type to a pointer
and forbids methods on it. `SchemaMethods()` consequently appends this
free-function registration (`internal/builder/gen_schema.go:748-751`), and the
template emits `func (Q) QSchema() ...`
(`internal/builder/schemas.go.tmpl:100-128`). The generated package then fails
to compile with the same `invalid receiver type Q (pointer or interface type)`
error reproduced directly above. The equivalent forwarding-interface form has
the same defect unless the base interface also happens to be present in
`Scan.Interfaces` under `Q`'s own name.

The same false classification bypasses the new safety paths: a
`NewJSONSchemaBuilder[Q]` entry is absent from
`InvalidReceiverBuilderRoots()`, while a `Declare(QSchema)` entry is absent
from `SchemaFreeFuncs()`, so the `RenderProviders()` and `--validate` guards
also do not fire (`internal/builder/gen_schema.go:778-804`;
`internal/builder/builder.go:71-94`). The tests cover only direct
`type PointerRoot *int` and direct interface declarations
(`internal/builder/validate_free_func_test.go:29-198`), leaving the forwarding
case unprotected even though the existing indirect-types fixture already
demonstrates that the generator resolves multi-level named definitions
(`internal/builder/testfixtures/indirecttypes/types.go:18-34`).

Impact: valid build-tagged input still produces uncompilable or silently
partial generated output in every path this branch's classifier is intended to
make safe. Classify with resolved Go type information (for example, the
`go/types` underlying type) rather than the immediate `dst` node, and add
free-function and builder-rejection fixtures for at least one forwarding
pointer definition; the resolved check should cover forwarding interfaces as
well.

## Outcome

material findings remain
