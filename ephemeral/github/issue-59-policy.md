Initial implementation is underway in parallel with #56. The helper-specific contract decisions are resolved:

- `Strict=true` requires every declared property and forces `additionalProperties=false` in both helper types, overriding explicit Required/AdditionalProperties settings. Non-strict behavior preserves explicit settings.
- Empty `EnumSchema` keeps the existing `SchemaNode` return signature and returns a node whose `MarshalJSON` reports a descriptive error; construction must not panic and encoding must not emit an invalid or permissive schema.
- Map-backed derived required keys are sorted; slice-backed ObjectSchema preserves insertion order.

These decisions unblock the full helper fix. #56 still carries the broader provider/API contract decision; that independent question does not require delaying helper implementation.
