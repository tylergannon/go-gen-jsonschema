# Adversarial review: jsonschema-tag build fixes, round 04

## Review target

The entire current branch `claude/gallant-austin-9fd4e3` at `8ad57f1`,
including the original three example repairs, the tagged-build gate, the new
free-function-root code-generation path, its TypeScript integration, and the
round-03 responses. I reviewed it against the same authoritative contracts as
the preceding rounds: repair the reported generation-tag compilation failures
without discarding supported registrations, retain the public `Declare`
free-function contract and its fluent options, prevent silent partial output,
and provide durable behavioral proof.

## Evidence inspected

- Repository `AGENTS.md`; `Declare` and `NewInterfaceImpl` public contracts;
  full `origin/main...HEAD` diff and all five branch commits
- The scanner's registration/type classification, builder schema mapping and
  root classification, Go template, TypeScript lowering, affected examples,
  fixture harness and generated goldens, CI workflow, and `justfile`
- `ephemeral/worklog/202609051649-fix-jsonschema-tag-build-errors.md` and review
  rounds 01 through 03
- Current GitHub issues #90, #91, and #92: #90 now has the corrected active-
  fixture premise, and #92 is closed with the explicit `--validate` rejection
- Baseline `go test ./...`: exit 0
- `just build-tagged`, `go build ./...`, `go vet ./...`,
  `go run ./internal/cmd/doc-gen -check`, and
  `git diff --check origin/main...HEAD`: exit 0

Round 03's two findings are resolved at their named surfaces. A non-rendered
named-pointer free-function root now fails `--validate` before output mutation,
with focused unit coverage, and issue #90 no longer asserts that the
indirect-types fixture is absent from `TestBasic`. The original three public
example packages also still compile under the generation tag, and the new
plain pointer-root accessor remains covered by generated-code golden and
runtime tests.

## Findings

### 1. issue — Legacy registered interface roots evade the new classifier and still generate an illegal method

`hasInvalidMethodReceiverBase` consults only `Scan.LocalNamedTypes` and returns
false when a name is absent (`internal/builder/gen_schema.go:714-724`). But the
scanner deliberately stores an interface registered by
`NewInterfaceImpl[I](...)` in `Scan.Interfaces` at
`internal/syntax/scan_result.go:397-410`, and then excludes that type from
`LocalNamedTypes` at lines 462-475. Consequently a valid free-function root
for that registered interface is omitted from `SchemaFreeFuncs()` and appended
to `SchemaMethods()` instead (`internal/builder/gen_schema.go:734-760`). The Go
template then emits `func (I) EntryPointName() ...` at
`internal/builder/schemas.go.tmpl:100-128`, which Go rejects because an
interface cannot be a receiver base. The round-04 `--validate` guard also does
not fire, because it depends on the same empty `SchemaFreeFuncs()` result
(`internal/builder/builder.go:71-74`).

A concrete source shape is a named interface `I`, an implementation registered
with `polytype.NewInterfaceImpl[I](Impl{})`, a
`func ISchema(I) json.RawMessage`, and `polytype.Declare(ISchema)`. Both marker
forms are expressly supported (`AGENTS.md:81-89`; `declare.go:12-20`), and the
builder already maps legacy registered interfaces as union schemas
(`internal/builder/gen_schema.go:812-844`). Yet generation routes this root to
an uncompilable method instead of preserving its free function.

Impact: the new helper and diagnostics claim to handle underlying "pointer or
interface" types, but the only added classification/validation tests use a
named pointer. A supported legacy interface registration still produces broken
generated Go and bypasses the new fail-fast validation contract. Classify names
present in `Scan.Interfaces` as invalid method receiver bases (or derive the
underlying kind from the package type information), then add a generated
fixture that registers a legacy interface through a free function and compiles
and calls the resulting accessor. Its `--validate` behavior should be asserted
too.

### 2. issue — `RenderProviders()` is silently broken on the new free-function-root path

The public `Declare` contract says method and free-function forms share the
same chained options, explicitly including `RenderProviders`
(`declare.go:12-17`). `NewForTypes` honors that option for both
`SchemaMethods` and `SchemaFuncs` by marking the receiver rendered
(`internal/builder/gen_schema.go:64-104`), and `writeSchema` therefore writes a
rendered root as `<Type>.json.tmpl` (`internal/builder/gen_schema.go:1268-1276`).
The new `SchemaFreeFuncs` template block, however, unconditionally reads
`<Type>.json` (`internal/builder/schemas.go.tmpl:131-143`). The only generated
`RenderedSchema` implementation is in a later block that ranges exclusively
over `SchemaMethods` (`internal/builder/schemas.go.tmpl:525-565`), so a rendered
named-pointer free-function root gets neither a usable schema accessor nor any
rendering entrypoint: calling its generated function panics because the
embedded file has the `.json.tmpl` suffix, and no API exists to execute the
template providers.

The new validation rejection compounds the inconsistency by rejecting every
`SchemaFreeFuncs()` root before checking `Rendered`: rendered/template types
are intentionally exempt from `ValidateJSON` because their schema depends on
runtime values (`AGENTS.md:98-100`), so the absence of a validator is not a
reason to reject that combination.

Impact: the branch turns a previously discarded root class into advertised
general free-function support but implements only the plain `.json` case. The
runtime test in `testfixtures/entrypoints` exercises no fluent option, so it
cannot catch this. Either generate a coherent free-function rendering surface
(including `.json.tmpl` access and provider execution), or reject
`RenderProviders()` on an invalid-receiver free-function root before writing
artifacts. The validation guard must reject only non-rendered roots that
actually require a generated validator.

## Outcome

material findings remain
