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

## Do not use `omitempty` for schema optionality

`omitempty` only controls Go marshaling. It does not remove an ordinary field
from the schema's `required` array. Combining an ordinary required field with
`omitempty` can produce JSON that fails its own generated schema when the Go
zero value is marshaled.

The legacy `jsonschema:"optional"` tag is no longer honored. Replace it with an
`Optional[T]` field and add `json:",omitzero"`.

## Validate before unmarshaling

Plain `json.Unmarshal` maps both a missing `Nullable[T]` property and an
explicitly null property to `Present == false`. Generate `ValidateJSON` and call
it before unmarshaling so missing required keys are rejected while explicit
null remains valid.

```go
if err := (Contact{}).ValidateJSON(data); err != nil {
    return err
}
var contact Contact
if err := json.Unmarshal(data, &contact); err != nil {
    return err
}
```

## OpenAI strict Structured Outputs

Strict schemas require every property to appear in `required`. Use
`Nullable[T]` for the required-plus-null pattern. Do not use `Optional[T]` in a
strict schema because it deliberately removes the property from `required`.

## Nullable enums and shared reference types

`Nullable` can wrap a registered enum or a struct registered with `AsRef()`:

```go
type Mode string

const (
    ModeFast Mode = "fast"
    ModeSafe Mode = "safe"
)

type Config struct {
    Mode   jsonschema.Nullable[Mode]   `json:"mode"`
    Shared jsonschema.Nullable[Shared] `json:"shared"`
}

var (
    _ = jsonschema.NewJSONSchemaMethod(Shared.Schema, jsonschema.AsRef())
    _ = jsonschema.NewJSONSchemaMethod(Config.Schema)
    _ = jsonschema.NewEnumType[Mode]()
)
```

Both properties remain in `required`. The enum renders as `anyOf` containing
its enum schema and null. The shared struct renders as `anyOf` containing its
`$ref` and null, while the referenced object remains available in `$defs`.

## Supported shapes

`Optional` supports scalars and named scalars, structs, pointers, arrays and
slices, explicit supported references, and registered interfaces. `Nullable`
supports scalars, registered enums, structs, pointers to structs, and structs
registered with `AsRef()`. Wrappers must be complete, direct named field types.
Aliases, defined wrappers, embedding, nesting, wrappers inside containers, and
unsupported Nullable shapes are rejected during generation.

See the compiling [`examples/optionality`](https://github.com/tylergannon/go-gen-jsonschema/tree/main/examples/optionality)
package for general wrapper coverage and
[`examples/ref_types`](https://github.com/tylergannon/go-gen-jsonschema/tree/main/examples/ref_types)
for nullable enum and `AsRef` validation coverage.
