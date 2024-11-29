# Valid schema descriptors

In this context, _renderable_ means that it can be compiled to the supported
subset of JSON Schema.

For a given type T, what makes it renderable?

* It can be ~int, ~string, or ~float.
* It can be a struct type whose every field is either (a) renderable or (b) ignored via `json:"-"`.
* It can be a type definition or type alias of a renderable type.
* It can be assigned type alternatives via `jsonschema.NewUnionType`.
* It can be a slice of one of the above.

## Special Cases

### Type Alternatives

### Enum

Enum types can be denoted by providing 