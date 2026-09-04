## Problem

Live probes on main 4376ce0 found:
- With type Count int, ConstSchema(Count(3), "") emits {"const":3,"type":"string"}; EnumSchema("", Count(1), Count(2)) likewise labels numeric values as strings.
- ObjectSchema.AddProperty with a quote in the key produces invalid JSON.
- Repeated marshaling of one strict JSONSchema produced four distinct encodings because required is built by ranging a map.
- JSONSchema.Strict preserves explicit AdditionalProperties=true, while ObjectSchema.Strict overrides it to false.
- EnumSchema[int]("") panics by indexing an empty argument list.

## Acceptance and proof

- Named integer/string values produce the correct schema type and accept their intended values under the schema validator.
- All legal property names, including quotes, backslashes, control characters, and Unicode, serialize and deserialize correctly.
- Repeated encoding of an unchanged schema is byte-stable, including required ordering.
- Choose and document one coherent strict-mode contract for both builders; tests exercise explicit AdditionalProperties values and required overrides.
- Define an intentional, documented empty-enum failure path instead of an index-out-of-range panic. Review any signature change before v1 freezes.
- Public-API regression cases exercise the actual marshaled schemas, not only internal type flags.

Strict/error policy is resolved in the implementation decision below; #59 is now unblocked independently of the remaining #56 provider decision. Implementation is in json_schema.go; focused regressions belong with its existing tests.

Implementation decision: Strict=true overrides explicit Required/AdditionalProperties in both helpers, requiring all declared properties and setting additionalProperties=false. Non-strict settings are preserved. Empty EnumSchema keeps the SchemaNode signature and returns a node whose MarshalJSON reports a descriptive error.
