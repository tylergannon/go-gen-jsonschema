# Generate JSON Schema

All `"type": "object"` schemas are given a unique property called
`__type` which is used as a discriminator.

## Valid schema descriptors

In this context, _renderable_ means that it can be compiled to the supported
subset of JSON Schema.

For a given type T, what makes it renderable?

* It can be ~int, ~string, or ~float.
* It can be a struct type whose every field is either (a) renderable or (b) ignored via `json:"-"`.
* It can be a type definition or type alias of a renderable type.
* It can be a slice of one of the above.
* It can be assigned type alternatives via `jsonschema.SetTypeAlternative`.

## Special Cases

### Type Alternatives

Type alternatives are denoted using `jsonschema.SetTypeAlternative`

### Enum

Enum types can be denoted by providing 

## Algorithm

For each named type:

1. If it provides type alternatives, recurse to those, and build a `UnmarshalJSON()` method.
2. If it is a slice type, recurse to the element type.
3. If it is a type definition or alias, recurse to the underlying type.
4. If it is a struct, process the struct type and each of its non-ignored fields..
5. If it is ~int, ~string, ~bool, or ~float, process it.
6. Return an error (unprocessable)

