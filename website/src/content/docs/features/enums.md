---
title: Enums
description: Generate JSON Schema enum values from Go constants.
---

## String constants

Register the containing schema with `WithEnum` for each enum field. Values are
discovered from typed constants in the same package.

```go
type Status string

const (
    StatusPending Status = "pending"
    StatusDone    Status = "done"
)

type Task struct {
    Status Status `json:"status"`
}

var _ = jsonschema.NewJSONSchemaMethod(
    Task.Schema,
    jsonschema.WithEnum(Task{}.Status),
)
```

## Integer and iota constants

Choose the wire representation for an integer-backed enum:

- `WithEnum` emits raw numeric constant values.
- `WithStringerEnum` emits constant names such as `LogDebug` and `LogInfo` as
  strings. It does not emit the return values of `String()`.

```go
type LogLevel int

const (
    LogDebug LogLevel = iota
    LogInfo
    LogError
)

type Config struct {
    LogLevel LogLevel `json:"logLevel"`
}

var _ = jsonschema.NewJSONSchemaMethod(
    Config.Schema,
    jsonschema.WithStringerEnum(Config{}.LogLevel),
)
```

Use `jsonschema.WithEnum(Config{}.LogLevel)` instead when the JSON contract
should contain numeric values.

## Encode and decode string mode

Generation adds one value `MarshalJSON` and pointer `UnmarshalJSON` to the
containing owner. These methods compose string-mode enum fields with any union
fields. They use constant identifiers, so `LogInfo` becomes `"LogInfo"` even
when a `String()` method returns different text. No global codec is added to
the enum type; another field registered with `WithEnum` remains numeric.

Supported adapted fields are direct integer-backed `E`, `Optional[E]`, and
`Nullable[E]`. Optional absence is omitted; Nullable null remains null. Present
values must match a declared constant. Decode errors leave the owner unchanged.
Validate external JSON first to enforce required fields and schema membership.

Unknown wire names and undeclared Go values are errors, including zero when
there is no zero-valued constant. Duplicate underlying values with different
names are ambiguous and rejected before generation writes files. Keep one
canonical constant name per value for string mode, or retain numeric mode.
Custom JSON hooks on adapted enum types and adapted pointers/slices or other
containers are rejected; move conversion to a supported named owner field.

In the example below, `ApplicationConfig` uses string mode while `Task`
intentionally uses numeric mode for the same enum types. Renaming constants
changes the string-mode wire contract and requires compatibility review.

The older package-level `NewEnumType[T]()` registration remains supported for
string enums, but field-level options make the containing schema's behavior
explicit and are preferred for new code.

See the compiling [`examples/stringer_enums`](https://github.com/tylergannon/go-gen-jsonschema/tree/main/examples/stringer_enums)
package for a complete example.
