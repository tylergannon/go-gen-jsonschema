## Problem

The shipped stringer_enums.Task schema expects constant-name strings such as "PriorityLow" and "LogInfo". Its Go fields marshal to 100 and 1, and json.Unmarshal rejects the schema's string values. Schema generation alone does not supply the advertised bidirectional boundary.

## Outcome

Provide encoding and decoding for the registered enum wire representation, using the same resolved mapping as schema generation.

## Acceptance and proof

- A generated consumer encodes integer-backed enums to their registered wire strings and decodes those strings back to the intended Go constants.
- Preserve per-field enum mode: the same Go type used with different registrations must not acquire an incorrect global codec.
- Resolve and document duplicate underlying values/aliases, unknown strings, unknown numeric values, zero values, and handwritten codec collisions according to #56.
- Preserve explicit constant-name behavior; do not silently switch to the result of String().
- Cover named integer types, ordinary string enums, supported Optional/Nullable enum fields, and supported enclosing/container contexts.
- Prove schema-valid encoding, semantic roundtrips, and expected rejection through generated code in a fresh consumer.
- Add the restored string-mode fixture from #33 to the relevant behavioral proof; keep generated examples and agent references consistent.

Depends on #56, #57, and #33. #57 establishes the shared containing-struct codec architecture; enum fixture/design work can proceed in parallel. Related: #46. Do not duplicate the existing fixture-restoration issue or silently broaden the supported enum/container set.