## Problem

On main 4376ce0, the shipped sealed_interface_slices example decodes {"events":[{"!kind":"Created","name":"first"}]} successfully, but json.Marshal emits {"events":[{"name":"first"}]}. Decoding that output fails because the discriminator is missing. The existing example test explicitly expects this lossy output.

## Outcome

Generate json.Marshaler implementations on containing structs so supported registered union values can be written and read using the same wire contract.

## Acceptance and proof

- Schema generation, marshaling, and unmarshaling consume the same resolved union registration, including legacy registration paths. Reject duplicate legacy-derived wire names before writing output; never suffix only the decoder mapping. Helper reuse must include field discriminator configuration, not merely the interface name.
- Encode the registered discriminator name and wire value for scalar I, supported Optional[I], and direct []I fields. Preserve ordinary field encoding and supported nesting.
- Demonstrate one concrete implementation used in separate fields with different discriminator names/values; encoding must respect each field's context.
- Cover explicit and legacy-derived wire values, single/multiple implementations, pointer/value implementations, and named discriminator keys that require JSON escaping.
- Return actionable errors for unregistered dynamic implementations and conflicting discriminator payloads; implement the contract's nil-interface and typed-nil behavior without panics.
- Detect handwritten MarshalJSON collisions before writing generated output; do not silently replace user code.
- A fresh consumer can encode, validate against the generated schema, decode, and recover semantically equivalent values. Valid input documents also survive decode/encode/decode.
- Replace the example's discriminator-dropping expectation, retain transactional decoder guarantees, and prove repeated generation produces no changes.

Depends on #56. Related: #18, #19, #47. New unsupported union container shapes and generic streaming are outside scope.