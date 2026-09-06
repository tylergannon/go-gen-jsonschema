# Plan: issues #86, #87, #88 in one session

Branch `claude/issues-86-87-88-plan-20be51`. Baseline `go test ./...` is green.
Sequence is fixed by dependency: #86 (enum marker) is independent, #87 (sealed
union inference) removes the field-level interface API, #88 (`SealedUnion`)
restores custom discriminators on top of #87. One commit per issue, one PR.

Gate after every step: `go test ./...`, `go generate ./...` in every
`examples/*` and `internal/builder/testfixtures/*` package, `go generate .`
(doc-gen for the skill), `just lint`, and `cd tests/typescript && npm ci && npm test`.

## Step 1: #86 enum marker `func (T) enum()`

Scanner (`internal/syntax`): after `loadPkgDecls`, walk `pkg.Types.Scope()`
for named types with an explicit method named `enum`; validate value receiver,
no params, no results, else diagnostic naming the type. For each valid marker
call `ResolveEnum` and store in `ScanResult.Constants` (the map `NewEnumType`
fills today). Zero constants is an error. Remote packages already get their own
`ScanResult`, so a marked type in another package is discovered the same way
(replaces `enumsremote` / `remoteenum` fixtures' `NewEnumType`).

Builder (`internal/builder/gen_schema.go`, `typegrammar.go`): field enum plans
currently come from `EnumV1[receiver][field]`. Add a second source: field type
resolves to a marked enum, mode = values. `.StringerEnum` keeps its path;
marked type plus `.StringerEnum` on the same field is a diagnostic.

Removals: `Declaration.Enum`, `WithEnum`, `NewEnumType`, `MarkerFuncNewEnumType`
parsing, `EnumType` marker struct, related scanner test cases.

Migrate: `examples/{enums,enums_stringmode?,iota_global,ref_types,self_contained,
template_rendering,test_options}`, builder fixtures `enums`, `traversal`,
`interfaces`, `v1_enums_stringmode` (if it uses `.Enum`), `tests/typescript/fixture`,
`internal/syntax/testfixtures/typescanner`. `iota_global` and `test_options`
do not build under the `jsonschema` tag today; the marker fixes `iota_global`,
`test_options` drops its nonexistent `WithDescription`.

Tests: golden regeneration of migrated fixtures; CLI negative tests (pointer
receiver, wrong signature, zero constants, marker + StringerEnum); marked type
with `String()` still emits values; round-trip codec test on a marked enum.

Docs: README enum section, `docs/tutorial.mdx`, `website/.../features/enums.md`,
`website/.../api/index.md`, `skills/polytype/SKILL.md` and
`references/registration-api.md`, `examples/README.md`; short migration note.

## Step 2: #87 sealed union inference, default discriminator only

Scanner: when type resolution reaches a named interface (local or remote),
inspect `types.Interface`: qualifies only if `NumExplicitMethods` includes an
unexported method; sealing method arriving via `NumEmbeddeds` is a diagnostic.
Enumerate package scope for named struct types; for each, find the sealing
method among `Named.Method(i)` (explicit methods only, so promoted methods are
excluded automatically), receiver kind gives value vs pointer variant, then
require `types.Implements(T or *T, iface)`. A struct that satisfies the
interface only via embedding gets the "excluded, inherited" diagnostic;
one whose explicit sealing method has a value receiver but only `*T` satisfies
the full interface gets the "invalid candidate" diagnostic. Populate the
existing `IfaceImplementations` (extend with pointer flags) so `mapInterface`,
`renderRegisteredInterfaceUnion`, the codec template, and the TypeScript
projection keep working. Zero variants is an error. Any reachable interface
field that is not sealed is an error at the field.

Builder: replace `IfaceV1[receiver][field]` with a per-interface table keyed by
`syntax.TypeID` (variants, pointer flags, discriminator = `"type"`, values =
type names). `registeredInterfaceField` reads from that table instead of field
config. Discriminator collision validation is unchanged.

Removals: `Declaration.Interface`, `WithInterface`, `WithInterfaceImpls`,
`WithDiscriminator`, `NewInterfaceImpl`, `Impl`, `Discriminator`,
`InterfaceOption`, `InterfaceOptionObj`, `InterfaceMarker`,
`parseInterfaceNestedOptions`, the legacy-vs-inline mixing checks.

Migrate: `examples/{interfaces_options,optionality,sealed_interface_slices,
uniontypes}`, `internal/builder/messages`, builder fixtures `interfaces`,
`union_codec`, `v1_interfaces_options`, `tests/typescript/fixture`,
`optionality/negative/nullable_interface`. Custom discriminators temporarily
become `"type"`; custom wire values (`"created"`, `"impl \"two\""`) become type
names. `union_codec` fixture variants that share a struct across several
fields with different wire names collapse to one union.

Tests: membership-drift golden on one representative union; CLI negatives for
each of the six diagnostics; value and pointer round trips; slices, Optional,
Nullable retained; determinism test still passes.

Docs: README interfaces section, tutorial, website `interfaces.md`, skill
files, wire-contract hazard note, migration note.

## Step 3: #88 `SealedUnion[I](name)`

Root package: `func SealedUnion[I any](discriminator string) SealedUnionMarker`.
Scanner: new `MarkerFuncSealedUnion`; type argument must resolve to the
current package (else diagnostic with location); string literal only (reuse
the existing literal check); duplicate per interface is a diagnostic; record
in `ScanResult.UnionDiscriminators[typeName]`. Builder: when building the
per-interface table, look up the interface's own package `ScanResult`; reject
non-interface and non-sealed targets; validate the property name with the
existing rule; run collision validation against it.

Migrate: restore `"!kind"` on `sealed_interface_slices`, `interfaces_options`,
`optionality`, and the TypeScript fixture (`"!kind"`, `"other-key"`).

Tests: CLI negatives for the seven listed diagnostics; multi-struct reuse of a
custom union; TypeScript narrowing on the custom property.

Docs: `SealedUnion` and the same-package rule in README, website, skill.

## Closeout

Rename worklog, regenerate the skill examples, final full gate, PR body
listing the three issues and the migration notes.
