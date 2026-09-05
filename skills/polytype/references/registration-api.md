# Registration API and CLI Reference

Read this when a schema type needs more than flat structs: enums, unions,
custom discriminators, free functions — or when you need exact CLI flags.

## Enums

### String enums — `Enum`

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
var _ = polytype.Declare(Task.Schema).
    Enum(Task{}.Status)
```

Produces `"status": {"type": "string", "enum": ["pending", "in_progress", "completed"]}`.

### Integer (iota) enums — `StringerEnum`

`StringerEnum` emits the **constant names** as string enum values
(`["LogDebug", "LogInfo", ...]`); plain `Enum` on an int type would emit
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

var _ = polytype.Declare(Config.Schema).
    StringerEnum(Config{}.LogLevel)
```

String mode generates codecs on the containing owner, composing with any
union fields. Marshal the owner value or pointer and decode into its pointer;
the same enum can still use numeric `Enum` in another field. Do not add a
global enum codec. Supported adapted fields are direct `E`, `Optional[E]`, and
`Nullable[E]`: absent/null wrappers bypass conversion, and present values use
constant identifiers. Unknown names/values, undeclared zero, ambiguous aliases,
custom enum JSON hooks, and other adapted containers are errors. Validate
external input before decoding for required-field and schema checks.
Registered enum fields cannot use `json:",string"`; generation rejects that
option before writing artifacts because its encoding differs from the schema.

`.Enum`/`.StringerEnum` are not a full replacement for the legacy
package-level `NewEnumType[T]()` when the enum type is shared across more
than one struct field: they only support a direct named enum, `Optional[E]`,
or `Nullable[E]` field, and annotating only some occurrences of a shared enum
type silently degrades the ones left unmarked (lost constraint, lost shared
TypeScript type). Keep a shared enum type on `NewEnumType[T]()`; it has no
fluent replacement.

## Discriminated unions (interface fields)

An interface-typed field becomes a union (`anyOf`) of its registered
implementations, discriminated by a `"type"` property. A direct
one-dimensional slice field (`[]PaymentMethod`) becomes an array whose `items`
contains that union. The generator emits owner `MarshalJSON` and `UnmarshalJSON` by default. Pass
`--formats=both` to add `go.yaml.in/yaml/v4` entry points for scalar values and
every slice element. YAML is translated into the JSON data model and decoded
through the same implementation. Both syntaxes use `type` as the default
discriminator property and JSON Schema property names are canonical. Go `yaml`
struct tags are ignored and nested custom `UnmarshalYAML` hooks are bypassed;
custom `UnmarshalJSON` hooks remain authoritative.

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
    ID      string          `json:"id"`
    Methods []PaymentMethod `json:"methods"`
}
```

Preferred per-field registration:

```go
// schema.go (//go:build jsonschema)
var _ = polytype.Declare(Payment.Schema).
    Interface(
        Payment{}.Methods,
        polytype.Discriminator("!kind"), // optional; default "type"
        polytype.Impl("credit_card", CreditCard{}),
        polytype.Impl("bank_transfer", BankTransfer{}),
    )
```

`Impl` binds each implementation to a stable wire discriminator used by both
the generated schema and owner encode/decode methods. Without explicit `Impl`
values, discriminators derive from Go type names.

The slice must be the direct field type. Fixed arrays, nested slices, named
slice containers, `Optional[[]I]`, and `Nullable[[]I]` are rejected during
generation. An `Optional[I]` scalar is supported; `Nullable[I]` is not.

Migration: `NewJSONSchemaMethod(Payment.Schema, WithInterface(Payment{}.Methods,
Impl(...), ...))` is now `Declare(Payment.Schema).Interface(Payment{}.Methods,
Impl(...), ...)`. `NewJSONSchemaMethod`/`NewJSONSchemaFunc` with `With*`
options, the split `WithInterface`/`WithInterfaceImpls`/`WithDiscriminator`
options, and the package-level `NewInterfaceImpl[I](impls...)` remain
supported and source-compatible; each carries a `Deprecated:` godoc comment
naming its fluent equivalent.

## Full registration surface

`Declare(fn)` is the entry point; `fn` is a method expression (`T.Schema`) or
a free function taking `T` as its sole parameter. Chain options onto the
returned `*Declaration[T]`:

- `.Accessor(field, T.method)` / `.Method(field, T.method)` / `.Function(field, fn)`
  — provider options: supply a field's schema at runtime instead of deriving
  it statically (see [`examples/providers_rendering`](../../../examples/providers_rendering)).
- `.Enum(field)` / `.StringerEnum(field)` — enum options.
- `.Interface(field, Discriminator(name), Impl(value, implementation), ...)` —
  sealed-interface options.
- `.Ref()` — render this type as `"$ref"` wherever it's referenced (see below).
- `.RenderProviders()` — generate `RenderedSchema()` and run providers at
  runtime (advanced; rendered types get no `ValidateJSON` because their
  schemas depend on runtime values).

`NewJSONSchemaMethod(T.Schema, ...opts)` / `NewJSONSchemaFunc(fn, ...opts)`
with their `With*` options, and the legacy `NewEnumType[T]()` /
`NewInterfaceImpl[I](impls...)`, remain supported for source compatibility.

Nested struct types are **inlined** into the parent schema (no `$ref`) by
default, so a shared Address struct appears in full wherever it is used —
unless that type is registered with `.Ref()`.

## Shared definitions (`$ref`/`$defs`) via `Ref`

Add `.Ref()` to a type's own registration to have it rendered as `"$ref":
"#/$defs/TypeName"` everywhere else it's referenced, instead of being inlined
at every call site. `$defs` are assembled per generated JSON file, keyed by
the type's bare name:

```go
// schema.go (//go:build jsonschema)
var _ = polytype.Declare(Shared.Schema).Ref()

var _ = polytype.Declare(Container.Schema) // references Shared
```

Notes:

- `.Ref()` only applies where `Shared` is referenced from *another*
  registered schema; `Shared`'s own top-level schema file is unaffected.
- Two distinct `.Ref()`'d types reachable in one generation run that share
  the same bare type name are a hard, generation-time error (`"AsRef
  definition name collision"`).
- Recursive/self-referencing `.Ref()`'d types are rejected, same as any
  other circular reference.

## Validation (`--validate`)

Opt in with `--validate` on generation and `new`. Each registered type gets
`ValidateJSON([]byte) error`; `--formats=both` also adds
`ValidateYAML([]byte) error`. Schemas are compiled once in `init()` using
`github.com/santhosh-tekuri/jsonschema/v6`. Failures return a
`*jsonschema.ValidationError` with `InstanceLocation` (path to the failing
field), `ErrorKind`, and nested `Causes`. Validation covers required fields,
types, unknown properties (rejected — `additionalProperties: false`), enum
membership, and nested structure. Validate LLM output *before* `json.Unmarshal`.

## CLI reference

```bash
polytype                 # same as `gen` in the current package
polytype gen [flags]
  -pretty            # indent the .json output
  -target DIR        # package to process (default: cwd)
  -no-changes        # fail (writing nothing) if schemas or requested TypeScript output would change
  -force             # rewrite even when unchanged; incompatible with -no-changes
  --validate         # generate validation methods for the selected formats
  --formats MODE     # decoding and validation: json (default) or both
  --typescript DIR   # generate structural TypeScript declarations in DIR
  --typescript-barrel # also generate index.ts type-only exports; requires --typescript
polytype new [flags]
  -out FILE          # stub file path ("" or "--" = stdout)
  -pkg NAME          # package name override (stdout mode)
  -methods 'T=Schema,U=Schema'   # required; one entry per type
  --validate         # include validation stubs for the selected formats
  --formats MODE     # validation stubs: json (default) or both
  --generate         # run `go generate ./...` in the target dir afterward
```

Environment: `JSONSCHEMA_NO_CHANGES` (any non-empty value) is equivalent to
`-no-changes` — useful in hooks/CI without editing `//go:generate` lines. It
checks requested TypeScript artifacts as well as schemas.

When installed via the go.mod tool directive, invoke everything as
`go tool polytype ...`.

## Generated layout

```
mypackage/
├── types.go            # your types + //go:generate directive
├── schema.go           # your stubs + registrations (//go:build jsonschema)
├── jsonschema/         # generated schema files, one per registered type
│   └── Person.json
├── jsonschema_gen.go   # generated implementations (//go:build !jsonschema)
└── web/src/generated/  # requested structural TypeScript output
    ├── types.ts
    └── index.ts        # only with --typescript-barrel
```

Generate validation and TypeScript declarations together with
`--validate --typescript web/src/generated`; add `--typescript-barrel` when an
`index.ts` type-only export is useful. `.Interface` and `.StringerEnum`
select the containing Go struct's JSON codecs automatically. TypeScript output
does not include a runtime decoder or validator: applications must validate
untrusted TypeScript-side data, and Go consumers should call generated
`ValidateJSON` before `json.Unmarshal`. Pin the tool and imported package to the
same explicit module release.

## Limitations and debugging

Not supported: map types, channels, functions, inline interfaces, recursive or
circular type references (detected and rejected), unsupported registered-
interface containers (fixed arrays, nested/named/optional/nullable slices), and
external package types other than `time.Time` (rendered as a string with RFC3339
guidance). Max nesting depth 100.

If generation fails:

1. Every type referenced in `schema.go` must exist in the package's Go source.
2. Check the build tag is exactly `//go:build jsonschema` on `schema.go`.
3. Look for circular references between types.
4. Enum consts must be declared in the same package as the enum type.
