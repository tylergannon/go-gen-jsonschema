# Registration API and CLI Reference

Read this when a schema type needs more than flat structs: enums, unions,
custom discriminators, free functions — or when you need exact CLI flags.

## Enums

### String enums — `WithEnum`

Values are auto-discovered from `const` declarations of the named type (the
consts must live in the same package as the type). No separate registration of
the enum type is needed:

```go
// types.go
type Status string

const (
    StatusPending    Status = "pending"
    StatusInProgress Status = "in_progress"
    StatusCompleted  Status = "completed"
)

type Task struct {
    ID     string `json:"id"`
    Status Status `json:"status"`
}
```

```go
// schema.go (//go:build jsonschema)
var _ = jsonschema.NewJSONSchemaMethod(
    Task.Schema,
    jsonschema.WithEnum(Task{}.Status),
)
```

Produces `"status": {"type": "string", "enum": ["pending", "in_progress", "completed"]}`.

### Integer (iota) enums — `WithStringerEnum`

`WithStringerEnum` emits the **constant names** as string enum values
(`["LogDebug", "LogInfo", ...]`); plain `WithEnum` on an int type would emit
the integers (`[0, 1, ...]`). Prefer the Stringer form for LLMs — names carry
meaning, integers don't:

```go
type LogLevel int

const (
    LogDebug LogLevel = iota
    LogInfo
    LogWarning
    LogError
)

var _ = jsonschema.NewJSONSchemaMethod(
    Config.Schema,
    jsonschema.WithStringerEnum(Config{}.LogLevel),
)
```

## Discriminated unions (interface fields)

An interface-typed field becomes a union (`anyOf`) of its registered
implementations, discriminated by a `"!type"` property. The generator also
emits an `UnmarshalJSON` on the containing struct that dispatches on the
discriminator. Arrays of interfaces (`[]PaymentMethod`) are **not** supported —
only a single interface field.

```go
// types.go
type PaymentMethod interface{ IsPaymentMethod() }

type CreditCard struct {
    Number string `json:"number"`
    Expiry string `json:"expiry"`
}
func (CreditCard) IsPaymentMethod() {}

type BankTransfer struct {
    AccountNumber string `json:"accountNumber"`
    RoutingNumber string `json:"routingNumber"`
}
func (BankTransfer) IsPaymentMethod() {}

type Payment struct {
    ID     string        `json:"id"`
    Method PaymentMethod `json:"method"`
}
```

Preferred per-field registration (v1 options):

```go
// schema.go (//go:build jsonschema)
var _ = jsonschema.NewJSONSchemaMethod(
    Payment.Schema,
    jsonschema.WithInterface(Payment{}.Method),
    jsonschema.WithInterfaceImpls(Payment{}.Method, CreditCard{}, BankTransfer{}),
    jsonschema.WithDiscriminator(Payment{}.Method, "!kind"), // optional; default "!type"
)
```

Legacy package-level registration (still works, but you cannot mix it with the
v1 per-field options in the same package):

```go
var _ = jsonschema.NewInterfaceImpl[PaymentMethod](CreditCard{}, BankTransfer{})
```

## Full registration surface

- `NewJSONSchemaMethod(T.Schema, ...opts)` — primary registration; one call per type.
- `NewJSONSchemaFunc(fn, ...opts)` — register a free function instead of a method.
- `NewEnumType[T]()` / `NewInterfaceImpl[I](impls...)` — legacy API; prefer the
  `With*` options.
- Options: `WithEnum(field)`, `WithStringerEnum(field)`, `WithInterface(field)`,
  `WithInterfaceImpls(field, impls...)`, `WithDiscriminator(field, name)`,
  `WithRenderProviders()` (runtime template rendering, advanced; rendered types
  get no `ValidateJSON` because their schemas depend on runtime values).

Nested struct types are **inlined** into the parent schema (no `$ref`), so a
shared Address struct appears in full wherever it is used.

## Validation (`--validate`)

Opt-in via the `--validate` flag on generation (and `--validate` on `new` so
the stubs include `ValidateJSON`). Each registered type gets
`ValidateJSON([]byte) error`; schemas are compiled once in `init()` using
`github.com/santhosh-tekuri/jsonschema/v6`. Failures return a
`*jsonschema.ValidationError` with `InstanceLocation` (path to the failing
field), `ErrorKind`, and nested `Causes`. Validation covers required fields,
types, unknown properties (rejected — `additionalProperties: false`), enum
membership, and nested structure. Validate LLM output *before* `json.Unmarshal`.

## CLI reference

```bash
gen-jsonschema                 # same as `gen` in the current package
gen-jsonschema gen [flags]
  -pretty            # indent the .json output
  -target DIR        # package to process (default: cwd)
  -no-changes        # fail (writing nothing) if regeneration would change any schema
  -force             # rewrite even when unchanged; incompatible with -no-changes
  --validate         # also generate ValidateJSON() methods
gen-jsonschema new [flags]
  -out FILE          # stub file path ("" or "--" = stdout)
  -pkg NAME          # package name override (stdout mode)
  -methods 'T=Schema,U=Schema'   # required; one entry per type
  --validate         # include ValidateJSON stubs
  --generate         # run `go generate ./...` in the target dir afterward
```

Environment: `JSONSCHEMA_NO_CHANGES` (any non-empty value) is equivalent to
`-no-changes` — useful in hooks/CI without editing `//go:generate` lines.

When installed via the go.mod tool directive, invoke everything as
`go tool gen-jsonschema ...`.

## Generated layout

```
mypackage/
├── types.go            # your types + //go:generate directive
├── schema.go           # your stubs + registrations (//go:build jsonschema)
├── jsonschema/         # generated schema files, one per registered type
│   └── Person.json
└── jsonschema_gen.go   # generated implementations (//go:build !jsonschema)
```

## Limitations and debugging

Not supported: map types, channels, functions, inline interfaces, recursive or
circular type references (detected and rejected), arrays of interface types,
external package types other than `time.Time` (rendered as a string with
RFC3339 guidance). Max nesting depth 100.

If generation fails:

1. Every type referenced in `schema.go` must exist in the package's Go source.
2. Check the build tag is exactly `//go:build jsonschema` on `schema.go`.
3. Look for circular references between types.
4. Enum consts must be declared in the same package as the enum type.
