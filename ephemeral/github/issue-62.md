## Outcome

Make the v1 portability promise executable against generated code, with coverage that distinguishes support for schemas, encoding, decoding, and validation.

## Acceptance and proof

- Drive a conformance suite from the capability matrix in #56; every supported row has positive/negative evidence and every exclusion has a clear failure or documented boundary.
- For supported valid Go values: encode, validate the generated schema, decode, and compare semantic meaning. For accepted input: decode/encode/decode preserves the declared meaning.
- Exercise absent/null/present-zero, pointer/value unions, field-specific discriminators (including legacy registration collisions), string-mode enums, container nil/empty semantics, named scalars, numeric boundaries, time values, refs, and supported custom-hook behavior. Include counterexamples/diagnostics for byte-like base64 slices, json:,string and unresolved external types so generic scalar/container claims cannot hide schema/codec disagreement.
- Invalid input and failed decode follow the declared mutation/transactionality contract.
- Exercise YAML through its canonical JSON semantics, including JSON property names, scalar resolution, unknown fields, and custom-hook boundaries. Do not claim YAML comment/format preservation or YAML output support without implementing it.
- Generate and execute fixtures in external consumer modules, with explicit execution of nested-module tests; a root go test invocation alone is not evidence that nested tests ran.
- Capture representative counterexamples and successful roundtrips as inspectable CI/release artifacts.

Depends on #56, #57, #58, and #59. Related: #33, #36. This is deterministic product conformance; real-agent task evaluations are tracked separately.