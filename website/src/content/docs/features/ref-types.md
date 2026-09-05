---
title: Shared definitions with Ref
description: Render a type once as a $ref into $defs instead of inlining it at every reference site.
---

By default, a struct type referenced from multiple places is **inlined** into
the schema at every reference site. Add the zero-arg `.Ref()` chain method to a
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

var _ = jsonschema.Declare(Shared.Schema).Ref()
var _ = jsonschema.Declare(Container.Schema)
```

`Container`'s generated schema gets a `$defs` object with one `Shared` entry,
and both `primary` and `others.items` reference it via `$ref` instead of
repeating its properties.

## Nullable references

A `.Ref()`-registered struct can be the value inside `Nullable[T]`:

```go
type NullableConfig struct {
    Shared jsonschema.Nullable[Shared] `json:"shared"`
}

var _ = jsonschema.Declare(Shared.Schema).Ref()
var _ = jsonschema.Declare(NullableConfig.Schema)
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

- `.Ref()` only changes how `Shared` is rendered at *other* types' reference
  sites; `Shared`'s own top-level schema file is unaffected.
- `$defs` are assembled per generated JSON file, keyed by the type's bare
  name. Two distinct `.Ref()`-registered types reachable in one generation
  run that share a bare name fail generation with a collision error.
- Recursive or self-referencing `.Ref()` types are rejected, the same as any
  other circular reference.

Migration: `NewJSONSchemaMethod(Shared.Schema, AsRef())` is now
`Declare(Shared.Schema).Ref()`. The legacy `AsRef()` option on
`NewJSONSchemaMethod`/`NewJSONSchemaFunc` remains supported and
source-compatible; it carries a `Deprecated:` godoc comment naming its
fluent equivalent.

See [`examples/ref_types`](https://github.com/tylergannon/go-gen-jsonschema/tree/main/examples/ref_types)
for the complete package, generated output, and validation tests.
