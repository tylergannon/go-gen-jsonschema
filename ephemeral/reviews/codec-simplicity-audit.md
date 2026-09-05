# Codec generator simplicity audit

Audit target: clean `codex/issue-58-enum-codec` at `8d8c7165826995826cd2a75bf01c51d6fa4115bd`. The paused `codex/v1-union-inspection` worktree was reference-only and remained untouched.

## Conclusion

The emitted codec architecture already matches the requested simple model. There is one value-receiver owner `MarshalJSON` and one pointer-receiver owner `UnmarshalJSON`; each declares a plain defined alias of the owner, embeds it in a wrapper, and shadows only registered union/enum fields with `json.RawMessage` (`internal/builder/schemas.go.tmpl:164-336`). Ordinary fields therefore stay with `encoding/json`. The mixed golden demonstrates union and enum fields sharing that one wrapper (`internal/builder/testfixtures/union_codec/jsonschema_gen.go.golden:96-181`). Sealed-union helpers are ordinary generated type switches (`schemas.go.tmpl:397-464`). The object/discriminator helper only checks and injects the sealed union discriminator (`schemas.go.tmpl:468-509`); it is not a generalized serializer.

No serializer redesign is warranted. One small generator-only simplification is safe and useful: make the template projection fresh and local to `RenderGoCode`, then delete the metadata that static use inspection proves dead. This should produce byte-identical generated artifacts.

## Current lowering and real duplication

Registration syntax is normalized into field configs in `NewForTypes` (`internal/builder/gen_schema.go:110-214`). `mapNamedType` then resolves enum fields and registered-interface fields before schema rendering (`gen_schema.go:978-1008`). Enum schema and codec generation already consume the same `EnumFieldPlan`: `resolveEnumFieldPlan` creates the entries once (`gen_schema.go:1048-1133`), `resolvedEnumField`/`schema` use them for JSON Schema (`gen_schema.go:1739-1763`), and `RenderGoCode` selects only plans whose wire representation needs adaptation (`gen_schema.go:1455-1475`). This is the desired shared owner-wrapper lowering and should remain.

Union schema and codec discovery call the same `resolveRegisteredInterfaceField` function (`gen_schema.go:1648-1668` and `2176-2216`). The repeated call is not currently divergent logic. The resulting union field plan is then projected once for the owner wrapper and once for deduplicated helper emission.

The unnecessary part is the lifetime of those template projections. `SchemaBuilder` stores `Imports`, `OwnerCodecs`, `YAMLTypes`, and `Interfaces` as mutable fields (`gen_schema.go:379-393`), although repository-wide use is confined to `RenderGoCode` and the template. `RenderGoCode` appends all three slices on every call (`gen_schema.go:1455-1535`). A second call on the same builder therefore duplicates owner, union-helper, and YAML declarations in the template input. These slices are derived views of `customTypes` and `enumFields`, not builder state needed by schema mapping.

There is also provably unread residue in that projection path:

- `registeredInterfaceField.FuncNameAlias` is computed and copied through `InterfaceProp.FuncNameAlias` (`gen_schema.go:1795-1803`, `1945-1953`, `2029-2038`, `2207-2216`), but helper names now come solely from `helperIdentity` and its hash (`gen_schema.go:2047-2082`).
- `EnumEntry.StringValue` and `EnumFieldPlan.Underlying` are written but never read after resolution (`gen_schema.go:289-312`, `1121-1133`, `1157-1171`). `WireName` is the consumed representation.
- `InterfaceOptionInfo.TypeName`, `InterfaceOptionInfo.PkgPath`, and `InterfaceInfo.PkgPath` are populated (`gen_schema.go:329-377`, `1494-1513`) but never referenced by `schemas.go.tmpl` or other Go code.
- `IsSpecialType` is unused (`gen_schema.go:493-501`). `sortedCustomTypeNames` is unused by production and retained only by its direct unit test (`gen_schema.go:1413-1420`; `internal/builder/gen_schema_determinism_test.go:10-16`).
- The commented abandoned flattening experiment inside `RenderGoCode` has no executable role (`gen_schema.go:1427-1453`).

## Machinery that remains load-bearing

- Keep the full union `helperIdentity` input and hashing (`gen_schema.go:2061-2082`). The same Go interface may have different field-specific discriminator properties, values, implementation sets, package identities, and pointer forms. Collapsing helpers by interface name would change wire behavior.
- Keep the owner collision and anonymous-embedding checks (`gen_schema.go:503-665`). They reject promoted handwritten methods, locally embedded generated owners, and generated codecs in foreign embedded types that are hidden by the generation build tag. The regression coverage at `internal/builder/owner_codec_test.go:14-153` directly enforces the user's exceptional-promotion rule and pre-write preservation.
- Keep `EmbeddedPath` access, nil guards, and pointer initialization (`gen_schema.go:2041-2145`). This is an explicit representation for supported promoted union fields, not generic serialization.
- Keep `Adapted` on `EnumFieldPlan` (`gen_schema.go:1107-1133`). Numeric-mode enums and string-backed enums already agree with ordinary `encoding/json`; only integer-backed string mode needs an owner shadow.
- Keep import alias allocation (`internal/builder/import_map.go:25-101`) and discriminator validation. They prevent generated identifier collisions and ambiguous sealed-union switches.

## Recommended bounded change

Build a fresh render view inside `RenderGoCode` from `customTypes` and `enumFields`, with empty `OwnerCodecs`, `Interfaces`, and `YAMLTypes`, and pass that value to the template. Do not append template-only projections back onto the long-lived builder. Remove the dead fields, dead helper/test, and commented block listed above in the same generator-only cleanup.

Proof should require the same builder instance to render twice with byte-identical output, all existing owner/union/enum focused tests, `go test ./...`, and two full regenerations with no generated-artifact diff. Because every removed datum has no consumer and the template inputs retain the same values and ordering, generated source and runtime behavior should remain unchanged.

## Baseline limitation

`GOFLAGS=-p=2 go test ./...` passed in the audit target. The reference union/inspection worktree failed its known Optional message assertion and also failed `internal/builder/TestBasic/test2-indirecttypes` under its pre-existing uncommitted `internal/builder/builder.go`; the task explicitly required preserving that paused integration edit, so this audit did not modify or diagnose it further.
