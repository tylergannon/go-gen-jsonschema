---
title: Optional and nullable fields
description: Choose whether a JSON property is required, omitted, or explicitly null.
---

Use the field type to express the JSON contract:

| Go field | JSON contract |
| --- | --- |
| `T` | Property is required and must be non-null. |
| `jsonschema.Optional[T]` | Property may be omitted and rejects JSON null. |
| `jsonschema.Nullable[T]` | Property is required and accepts JSON null. |

```go
type Contact struct {
    Name  string                      `json:"name"`
    Email jsonschema.Optional[string] `json:"email,omitzero"`
    Phone jsonschema.Nullable[string] `json:"phone"`
}
```

`Optional[T]` fields must use `json:",omitzero"`. Both wrappers preserve
present zero and empty values through their `Present` and `Value` fields.

## Validate before unmarshaling

Plain `json.Unmarshal` cannot distinguish a missing `Nullable[T]` property from
an explicitly null property. When required-key presence matters, generate
`ValidateJSON` and call it before unmarshaling.

## OpenAI strict Structured Outputs

Strict schemas require every property to appear in `required`. Use
`Nullable[T]` for the required-plus-null pattern. Do not use `Optional[T]` in a
strict schema because it deliberately removes the property from `required`.

## Supported shapes

`Optional` supports scalars and named scalars, structs, pointers, arrays and
slices, explicit supported references, and registered interfaces. `Nullable`
supports scalars, structs, and pointers to structs. Wrappers must be complete,
direct named field types; aliases, embedding, and nested wrappers are rejected.

See the compiling [`examples/optionality`](https://github.com/tylergannon/go-gen-jsonschema/tree/main/examples/optionality)
package for generated output and negative cases.
