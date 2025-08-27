# go-gen-jsonschema Examples

This directory contains comprehensive examples demonstrating all features of go-gen-jsonschema. Each subdirectory showcases different capabilities of the tool.

## Quick Start

Each example contains:
- `types.go` - Go type definitions
- `schema.go` - Schema registration with options
- Generated `jsonschema_gen.go` and `jsonschema/` directory after running `go generate`

To generate schemas:
```bash
cd [example-directory]
go generate
```

## Examples Overview

### Basic Examples

#### `basictypes/`
Demonstrates fundamental schema generation for simple types.
- Basic Go types (int, string, float, bool)
- Simple structs with documentation
- Nested structs
- Optional fields with `omitempty`

#### `structs/`
Complex struct examples with real-world patterns.
- Embedded structs
- Time.Time fields with RFC3339 format guidance
- Nested struct composition
- Field documentation
- Comprehensive test coverage

#### `indirecttypes/`
Shows handling of type aliases and indirection.
- Type aliases
- Pointer types
- Custom type definitions

### Enum Examples

#### `enums/`
String-based enums with explicit values.
- String-typed constants
- Auto-detection of string enum values
- Multiple enums in one package

#### `enums_stringmode/`
Alternative enum handling with string mode.
- String representation of numeric enums
- V1 API enum configuration

#### `stringer_enums/`
Integer enums with fmt.Stringer implementation.
- Iota-based constants
- Custom String() methods
- Enum validation

#### `iota_global/`
Global iota enum example.
- Package-level iota constants
- Enum value detection

### Interface & Union Types

#### `uniontypes/`
Discriminated union types using Go interfaces.
- Interface-based type unions
- Multiple implementations
- Discriminator properties
- Note: Arrays of interfaces (`[]Interface`) are not yet supported

#### `interfaces_options/`
Advanced interface configuration with V1 API.
- Custom interface implementations
- Discriminator configuration
- Implementation registration

### Provider & Template Examples

#### `providers_rendering/`
Template-based schema generation with runtime providers.
- Field-level schema providers
- Runtime template rendering
- Three provider types:
  - `WithStructAccessorMethod` - No-arg struct method
  - `WithStructFunctionMethod` - Struct method with field arg  
  - `WithFunction` - Free function provider

#### `template_rendering/`
Basic template rendering example.
- Template-based schema generation
- Static templates

#### `self_contained/`
Self-contained schema generation example.
- Complete example in single package
- No external dependencies

### Test & Configuration

#### `test_options/`
Various configuration options and edge cases.
- Different registration patterns
- Configuration options testing

## Key Patterns

### Basic Schema Registration
```go
func (MyType) Schema() json.RawMessage { 
    panic("not implemented") 
}
var _ = jsonschema.NewJSONSchemaMethod(MyType.Schema)
```

### Enum Registration
```go
type Status string
const (
    StatusPending Status = "pending"
    StatusActive  Status = "active"
)
var _ = jsonschema.NewEnumType[Status]()
```

### Interface/Union Type Registration
```go
type Shape interface{ /* methods */ }
type Circle struct{ /* fields */ }
type Rectangle struct{ /* fields */ }

var _ = jsonschema.NewInterfaceImpl[Shape](
    Circle{},
    Rectangle{},
)
```

### Provider-Based Schema Generation
```go
var _ = jsonschema.NewJSONSchemaMethod(
    Example.Schema,
    jsonschema.WithStructAccessorMethod(Example{}.A, (Example).ASchema),
    jsonschema.WithStructFunctionMethod(Example{}.B, (Example).BSchema),
    jsonschema.WithFunction(Example{}.C, BoolSchema),
    jsonschema.WithRenderProviders(),
)
```

## Running All Examples

To test all examples:
```bash
for dir in */; do 
    echo "Testing $dir"
    (cd "$dir" && go generate)
done
```

## Notes

- All examples use build tags to separate schema registration from normal builds
- The `//go:build jsonschema` tag is used in `schema.go` files
- Generated code uses `//go:build !jsonschema` to exclude from schema generation
- External types like `time.Time` are handled with descriptive guidance for LLMs