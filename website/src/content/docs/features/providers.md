---
title: Provider-rendered schemas
description: Supply field schemas at runtime when static generation is not enough.
---

Provider hooks replace selected field schemas with `json.Marshaler` values from
functions or methods:

```go
func BoolSchema(_ bool) json.Marshaler {
    return json.RawMessage(`{"type":"boolean"}`)
}

var _ = jsonschema.NewJSONSchemaMethod(
    Config.Schema,
    jsonschema.WithFunction(Config{}.Enabled, BoolSchema),
    jsonschema.WithRenderProviders(),
)
```

Available provider options are:

- `WithFunction(field, fn)` for a package function;
- `WithStructAccessorMethod(field, method)` for a receiver method;
- `WithStructFunctionMethod(field, method)` for a receiver method that accepts the field value;
- `WithRenderProviders()` to generate `RenderedSchema()` and execute providers at runtime.

Provider implementations must be available in normal builds because
`RenderedSchema()` calls them at runtime. A rendered type does not receive
`ValidateJSON()` because its schema depends on runtime values.

See [`examples/providers_rendering`](https://github.com/tylergannon/go-gen-jsonschema/tree/main/examples/providers_rendering)
for all three provider shapes and a runtime test.
