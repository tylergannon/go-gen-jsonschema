---
title: Shared definitions with AsRef
description: Render a type once as a $ref into $defs instead of inlining it at every reference site.
---

By default, a struct type referenced from multiple places is **inlined** into
the schema at every reference site. Add the zero-arg `AsRef()` option to a
type's own registration to render it once as `"$ref": "#/$defs/TypeName"`
wherever another registered schema references it instead:

```go
type Shared struct {
    Name string `json:"name"`
}

type Container struct {
    Primary Shared   `json:"primary"`
    Others  []Shared `json:"others"`
}

var _ = jsonschema.NewJSONSchemaMethod(Shared.Schema, jsonschema.AsRef())
var _ = jsonschema.NewJSONSchemaMethod(Container.Schema)
```

`Container`'s generated schema gets a `$defs` object with one `Shared` entry,
and both `primary` and `others.items` reference it via `$ref` instead of
repeating its properties.

## Nullable references

An `AsRef()` struct can be the value inside `Nullable[T]`:

```go
type NullableConfig struct {
    Shared jsonschema.Nullable[Shared] `json:"shared"`
}

var _ = jsonschema.NewJSONSchemaMethod(Shared.Schema, jsonschema.AsRef())
var _ = jsonschema.NewJSONSchemaMethod(NullableConfig.Schema)
```

The property remains required and accepts either the shared definition or JSON
null. Its generated shape is equivalent to:

```json
{
  "anyOf": [
    { "$ref": "#/$defs/Shared" },
    { "type": "null" }
  ]
}
```

The generator keeps the reachable `Shared` definition in the containing
schema's `$defs`; nullable wrapping does not inline or drop it.

Notes:

- `AsRef()` only changes how `Shared` is rendered at *other* types' reference
  sites; `Shared`'s own top-level schema file is unaffected.
- `$defs` are assembled per generated JSON file, keyed by the type's bare
  name. Two distinct `AsRef()`-registered types reachable in one generation
  run that share a bare name fail generation with a collision error.
- Recursive or self-referencing `AsRef()` types are rejected, the same as any
  other circular reference.

See [`examples/ref_types`](https://github.com/tylergannon/go-gen-jsonschema/tree/main/examples/ref_types)
for the complete package, generated output, and validation tests.
