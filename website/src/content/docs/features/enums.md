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

Use `WithStringerEnum` when an integer-backed enum should appear as meaningful
strings. The schema emits constant names such as `LogDebug` and `LogInfo`, not
raw integers or the return values of `String()`.

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

The older package-level `NewEnumType[T]()` registration remains supported for
string enums, but field-level options make the containing schema's behavior
explicit and are preferred for new code.

See the compiling [`examples/stringer_enums`](https://github.com/tylergannon/go-gen-jsonschema/tree/main/examples/stringer_enums)
package for a complete example.
