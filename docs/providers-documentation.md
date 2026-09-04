# Provider-rendered schemas

Provider options customize selected field schemas when the schema depends on a
runtime value. They are part of the retained v1 provider API. Providers change
schema rendering; they do not create a codec for the field's Go value.

## Registration

Use one of these options on a schema registration:

```go
var _ = jsonschema.NewJSONSchemaMethod(
    Example.Schema,
    jsonschema.WithStructAccessorMethod(Example{}.A, (Example).ASchema),
    jsonschema.WithStructFunctionMethod(Example{}.B, (Example).BSchema),
    jsonschema.WithFunction(Example{}.C, BoolSchema),
    jsonschema.WithRenderProviders(),
)
```

- `WithStructAccessorMethod(field, method)` calls a receiver method that needs
  only the containing value.
- `WithStructFunctionMethod(field, method)` calls a receiver method with the
  selected field value.
- `WithFunction(field, fn)` calls a package function with the selected field
  value.
- `WithRenderProviders()` enables runtime rendering of the provider template.

Each provider returns a `json.Marshaler` containing the schema fragment for its
field. The provider implementation must be available in normal builds because
the generated `RenderedSchema()` method invokes it at runtime.

## Generated output and limits

When providers are configured, generation writes a `jsonschema/<Type>.json.tmpl`
file. `Schema()` returns that template as raw bytes. With
`WithRenderProviders()`, the generated type also has:

```go
RenderedSchema() (json.RawMessage, error)
```

`RenderedSchema()` executes providers, substitutes their JSON fragments, and
returns the rendered schema. The repository's
[`examples/providers_rendering`](../examples/providers_rendering) package
exercises accessor, receiver-function, and free-function providers at runtime.

A provider-rendered schema is runtime-dependent, so generated static
`ValidateJSON` methods are not emitted for that type. Use the returned schema
with the validator or consumer that owns the runtime request. Provider options
also do not add generated `MarshalJSON` methods for ordinary values or union
implementations; encoding remains the responsibility of the value's normal Go
JSON methods.

The provider API is intentionally retained for v1. A future typed-props API may
be explored additively after v1; it does not replace these functioning options.
