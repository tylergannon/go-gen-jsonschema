# Issue #63 support inventory

Snapshot: 2026-09-04, branch `codex/issue-63-cleanup` including the parent integration of the no-op CLI flag removal. This is an inspectable audit,
not a replacement for the accepted contract in `docs/spec/v1.md`.

## Exported API classification

| Surface | Evidence | Classification |
| --- | --- | --- |
| `DataType`, `SchemaNode`, `JSONUnionType`, `SchemaProperty`, `ParentSchema`, `ObjectSchema`, `JSONSchema` | `json_schema.go` | Supported manual schema model and JSON marshaling. These methods marshal schema documents; they are not generated codecs for user values. |
| `StringSchema`, `BoolSchema`, `IntSchema`, `ArraySchema`, `EnumSchema`, `ConstSchema`, `RefSchemaEl`, `UnionSchemaEl` | `json_schema.go` | Supported manual schema helpers; `ObjectSchema` preserves insertion order and `JSONSchema.Properties` is map-backed/sorted. |
| `Optional[T]` | `optionality.go`, `optionality_test.go`, `examples/optionality` | Supported value wrapper for a direct named struct field that may be absent but rejects JSON `null`; zero is absent, `Present` distinguishes present zero/empty values, and `json:",omitzero"` is required. |
| `Nullable[T]` | `optionality.go`, `optionality_test.go`, `examples/optionality`, `examples/ref_types` | Supported value wrapper for a required field that accepts JSON `null`; zero encodes as `null`, `Present` distinguishes a non-null value, and generated validation is needed to distinguish missing from explicit `null`. |
| `NewJSONSchemaMethod` | `union_type.go`, generated examples | Primary supported registration marker. |
| `NewJSONSchemaFunc` | `union_type.go`, `internal/builder/testfixtures/entrypoints` | Supported free-function registration marker. |
| `NewJSONSchemaBuilder` | `union_type.go`, `internal/builder/testfixtures/entrypoints/schema.go`, `internal/builder/testfixtures/entrypoints/jsonschema_gen.go.golden` | Supported builder-style registration marker; the generated method is `BuilderTypeSchema()` and reads `jsonschema/BuilderType.json`. The entrypoint golden list now checks this schema explicitly; `providers_builder` covers provider rendering through the method marker, not this builder marker. |
| `NewEnumType`, `NewInterfaceImpl` | `union_type.go`, `examples/enums`, `examples/uniontypes` | Retained legacy registration markers; field-level options are preferred for new code. |
| `WithEnum`, `WithStringerEnum`, `WithInterface*`, `Discriminator`, `Impl`, `AsRef` | `union_type.go`, current examples | Supported generation options with the limits in the accepted v1 matrix. |
| `WithFunction`, `WithStructAccessorMethod`, `WithStructFunctionMethod`, `WithRenderProviders` | `union_type.go`, `examples/providers_rendering` | Retained provider API; runtime schema rendering only. Provider-rendered types do not receive static validation. |
| `Tool`, `BuildTool`, `Tool0`, `Tool1`, `SomeCoolTool` | deleted `tool_types.go` | Removed unfinished `nobuild` experiment; no shipped API or example depended on it. |

## Generated methods and codec boundary

| Generation condition | Generated surface | Classification |
| --- | --- | --- |
| Every registered method/function/builder | Generated accessor using the registered method/function name; entrypoint fixture emits `Schema`, `FuncTypeSchema`, and `BuilderTypeSchema` | Supported schema accessor. `NewJSONSchemaBuilder[T](fn)` uses the builder function's name for the generated accessor and generated schema file; it does not call `fn` as a runtime codec. |
| `--validate` | `ValidateJSON([]byte) error` | Supported opt-in schema validation. Validation is separate from typed decoding. |
| `--formats=both` | YAML entry points and `ValidateYAML` when validation is enabled | Supported opt-in canonical YAML-to-JSON path. |
| Registered interfaces | Owner-side `UnmarshalJSON`; YAML counterpart when enabled | Supported selected union decoding, preserving documented value/pointer forms. |
| `WithRenderProviders()` | `RenderedSchema() (json.RawMessage, error)` | Supported runtime schema rendering; no static validator for runtime-dependent output. |
| Concrete union/enum value encoding | No generated discriminator-aware `MarshalJSON` | Intentionally not promised. Implementations that must round-trip union JSON provide their own encoder. |

The generated surface does not imply that schema-valid JSON can always be
encoded or decoded for arbitrary Go types. The accepted contract separates
schema generation, validation, JSON decoding, YAML decoding, and value
encoding, and excludes unsupported mappings explicitly.

## CLI flags

| Flag | Current evidence on base | #63 target classification |
| --- | --- | --- |
| `-pretty`, `-target`, `-no-changes`, `-force` | `gen-jsonschema/main.go` | Supported generation controls. |
| `--validate`, `--formats` | `gen-jsonschema/main.go` | Supported selected validation/decoding controls. |
| `new -out`, `-pkg`, `-methods`, `--validate`, `--formats`, `--generate` | `gen-jsonschema/main.go` | Supported scaffold controls. |
| `-num-test-samples` | `gen-jsonschema/main.go` on the base snapshot; it has no effect | Remove before API freeze; the flag, internal fields, and documentation are removed together; executable help and rejection checks pass. |

## Shipped examples

| Example | Evidence and classification |
| --- | --- |
| `basictypes`, `structs`, `indirecttypes` | Compiling schema-generation examples for ordinary Go shapes. |
| `enums`, `enums_stringmode`, `stringer_enums`, `iota_global`, `self_contained`, `test_options` | Enum registration and constant-mode examples; current support is field-specific where documented. |
| `interfaces_options`, `uniontypes`, `sealed_interface_slices` | Interface registration, discriminators, value/pointer implementations, and selected union decoding. Concrete encoding still owns discriminator emission. |
| `optionality` | Optional/Nullable positive and negative shape proof, including generated validation and runtime wrapper behavior. |
| `ref_types` | `AsRef()`/`$defs` and nullable reference validation proof. |
| `providers_rendering`, `template_rendering` | Runtime provider/template schema rendering. These are schema-rendering examples, not general codecs. |

## Related issue dispositions

| Issue | Factual disposition for parent | Action here |
| --- | --- | --- |
| #10 | Open. Several checklist items concern generated `UnmarshalJSON` ownership and cross-package behavior; current selected local-owner coverage does not establish all of that issue's remaining scope. | Leave open; no issue mutation. |
| #17 | Open but superseded by the explicit `Optional[T]`/`Nullable[T]` contract in #32; its legacy `jsonschema:"optional"` plus null proposal is not current support. | Close as superseded by #32; do not advertise the legacy tag. |
| #28 | Open but superseded by closed #32; its primitive-only and deprecation proposal is obsolete. | Close as superseded by #32; route users to the current wrapper contract. |
| #32 | Closed. Its explicit wrapper decisions and proof requirements are the current historical source for Optional/Nullable behavior. | Treat as resolved provenance; current v1 limits remain in `docs/spec/v1.md`. |

## Checks for this slice

- `GOFLAGS=-p=2 go test ./...` passed before editing.
- Focused stale-claim review passed for owned current documentation.
- `TestBasic/test7-entrypoints` includes the builder schema golden and passed.
- Final `GOFLAGS=-p=2 go test ./...` passed after removing the no-op flag.
- Built CLI help omits the removed flag and rejects it with exit 2; see `cli-proof.txt`.
- `npm ci` and `npm run check` passed using the committed website lockfile.
