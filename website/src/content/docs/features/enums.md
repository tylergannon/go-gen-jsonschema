---
title: Enums
description: Generate JSON Schema enum values from Go constants.
---

## String constants

Register the containing schema with `.Enum` for each enum field. Values are
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

var _ = polytype.Declare(Task.Schema).
    Enum(Task{}.Status)
```

## Integer and iota constants

Choose the wire representation for an integer-backed enum:

- `.Enum` emits raw numeric constant values.
- `.StringerEnum` emits constant names such as `LogDebug` and `LogInfo` as
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

var _ = polytype.Declare(Config.Schema).
    StringerEnum(Config{}.LogLevel)
```

Use `.Enum(Config{}.LogLevel)` instead when the JSON contract should contain
numeric values.

## Encode and decode string mode

Generation adds one value `MarshalJSON` and pointer `UnmarshalJSON` to the
containing owner. These methods compose string-mode enum fields with any union
fields. They use constant identifiers, so `LogInfo` becomes `"LogInfo"` even
when a `String()` method returns different text. No global codec is added to
the enum type; another field registered with `.Enum` remains numeric.

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
Registered enum fields cannot use `json:",string"`; generation rejects that
option before writing artifacts because its encoding differs from the schema.

In the example below, `ApplicationConfig` uses string mode while `Task`
intentionally uses numeric mode for the same enum types. Renaming constants
changes the string-mode wire contract and requires compatibility review.

Migration: `NewJSONSchemaMethod(Task.Schema, WithEnum(Task{}.Status))` is now
`Declare(Task.Schema).Enum(Task{}.Status)`. The legacy `NewJSONSchemaMethod`/
`NewJSONSchemaFunc` with `WithEnum`/`WithStringerEnum`, and the older
package-level `NewEnumType[T]()`, remain supported and source-compatible;
each carries a `Deprecated:` godoc comment naming its fluent equivalent.

Field-level `.Enum`/`.StringerEnum` is not a full replacement for
`NewEnumType[T]()` when the enum type is shared across more than one struct
field: only a direct named enum, `Optional[E]`, or `Nullable[E]` field is
supported at the field level, and marking only some occurrences of a shared
enum type silently degrades the ones left unmarked (they lose their
constraint and their shared TypeScript type). Keep a shared enum type on the
package-level `NewEnumType[T]()` form; it has no fluent replacement.

See the compiling [`examples/stringer_enums`](https://github.com/tylergannon/polytype/tree/main/examples/stringer_enums)
package for a complete example.
