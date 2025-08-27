# go-gen-jsonschema Examples

This directory contains comprehensive examples demonstrating all features of go-gen-jsonschema. Each subdirectory showcases different capabilities of the tool.

## Quick Start

Each example contains:
- `types.go` - Go type definitions
- `schema.go` - Schema registration with options

To generate schemas:
```bash
cd [example-directory]
go generate ./...
```

## Core Features by Example

### 1. `basictypes/` - Basic Schema Generation
Demonstrates fundamental schema generation for simple types.

**Features:**
- Basic Go types (int, string, float)
- Simple structs with documentation
- Nested structs
- Optional fields with `omitempty`

**Key Pattern:**
```go
func (SimpleStruct) Schema() json.RawMessage { panic("not implemented") }
var _ = jsonschema.NewJSONSchemaMethod(SimpleStruct.Schema)
```

### 2. `enums/` - String-Based Enums
Shows string constant enums with explicit values.

**Features:**
- String-typed constants (`const Status string = "value"`)
- Auto-detection of string enum values
- Multiple enums in one package
- Using enums in structs

**Key Pattern:**
```go
type Status string
const (
    StatusPending Status = "pending"
    StatusActive  Status = "active"
)
var _ = jsonschema.NewEnumType[Status]()
```

### 3. `test_options/enum_iota_types.go` - Iota Enums
Demonstrates integer-based enums using iota.

**Features:**
- Iota-based enums (`const X Type = iota`)
- Custom integer values
- Stringer implementations
- Integer enum schema generation

**Key Pattern:**
```go
type Color int
const (
    ColorRed Color = iota
    ColorGreen
    ColorBlue
)
```

### 4. `stringer_enums/` - Stringer Enums with String Schema
Shows how to generate string schemas for integer enums with Stringer.

**Features:**
- Integer enums with String() methods
- `WithStringerEnum` option for string representation
- Contrast with regular `WithEnum` (integer values)

**Key Pattern:**
```go
func (l LogLevel) String() string { /* ... */ }

var _ = jsonschema.NewJSONSchemaMethod(
    Config.Schema,
    jsonschema.WithStringerEnum(Config{}.LogLevel), // Uses String() values
)
```

### 5. `v1/providers_rendering/` - Template Providers
Demonstrates schema templates with runtime field providers.

**Features:**
- Struct method providers (`WithStructAccessorMethod`)
- Function method providers (`WithStructFunctionMethod`)
- Free function providers (`WithFunction`)
- Runtime schema rendering (`WithRenderProviders`)

**Key Patterns:**
```go
// Provider implementations
func (Example) ASchema() json.Marshaler { 
    return json.RawMessage(`{"type":"string"}`) 
}
func (Example) BSchema(_ int) json.Marshaler { 
    return json.RawMessage(`{"type":"integer"}`) 
}
func BoolSchema(_ bool) json.Marshaler { 
    return json.RawMessage(`{"type":"boolean"}`) 
}

// Registration
var _ = jsonschema.NewJSONSchemaMethod(
    Example.Schema,
    jsonschema.WithStructAccessorMethod(Example{}.A, (Example).ASchema),
    jsonschema.WithStructFunctionMethod(Example{}.B, (Example).BSchema),
    jsonschema.WithFunction(Example{}.C, BoolSchema),
    jsonschema.WithRenderProviders(),
)
```

This generates:
- `Schema()` - Returns template with `{{.a}}`, `{{.b}}`, `{{.c}}` placeholders
- `RenderedSchema()` - Executes providers and fills template at runtime

### 6. `structs/` - Complex Structures
Advanced struct patterns and relationships.

**Features:**
- Embedded structs
- Recursive types (trees, linked lists)
- Maps and slices
- Time fields
- Complex nesting

### 7. `uniontypes/` - Interface Unions
Discriminated unions using Go interfaces.

**Features:**
- Interface-based unions
- Multiple implementations
- Discriminator property (`!type` by default)
- Custom discriminators with `WithDiscriminator`

**Key Pattern:**
```go
type Shape interface{ isShape() }
type Circle struct{ Radius float64 }
type Square struct{ Side float64 }

var _ = jsonschema.NewInterfaceImpl[Shape](Circle{}, Square{})
```

### 8. `indirecttypes/` - Pointers and References
Various forms of type indirection.

**Features:**
- Pointer types
- Slices and arrays
- Maps with complex values
- Named type aliases

## V1 Options API

The v1 API provides consolidated configuration through options:

### Enum Options
- `WithEnum(Field)` - Mark field as enum (auto-detects string/int)
- `WithStringerEnum(Field)` - Use String() method for enum values
- `WithEnumName(Value, "name")` - Override specific enum value names

### Interface Options
- `WithInterface(Field)` - Mark field as discriminated union
- `WithInterfaceImpls(Field, Impl1{}, Impl2{})` - Explicit implementations
- `WithDiscriminator(Field, "!kind")` - Custom discriminator property

### Provider Options
- `WithStructAccessorMethod(Field, Method)` - No-arg struct method
- `WithStructFunctionMethod(Field, Method)` - Struct method with field arg
- `WithFunction(Field, Func)` - Free function provider
- `WithRenderProviders()` - Generate RenderedSchema() method

## Alternative Registration Forms

### Method Form (Traditional)
```go
func (T) Schema() json.RawMessage { panic("not implemented") }
var _ = jsonschema.NewJSONSchemaMethod(T.Schema, /* options */)
```

### Function Form
```go
func Schema(T) json.RawMessage { panic("not implemented") }
var _ = jsonschema.NewJSONSchemaFunc(Schema, /* options */)
```

### Builder Form
```go
func Schema() json.RawMessage { panic("not implemented") }
var _ = jsonschema.NewJSONSchemaBuilder(Schema, /* options */)
```

## Build Tags

Schema files use build tags to exclude them from normal builds:
```go
//go:build jsonschema
```

## Generated Output

Running `go generate` creates:
- `jsonschema/*.json` - Static schema files
- `jsonschema/*.json.tmpl` - Template schemas (with providers)
- `jsonschema_gen.go` - Go code for runtime schema access

The generated code includes:
- `Schema()` methods returning embedded JSON
- `RenderedSchema()` methods (when providers are used)
- Unmarshaler helpers for discriminated unions

## Testing Your Schemas

Schemas can be validated and tested:
```go
import "encoding/json"

func TestSchema(t *testing.T) {
    var s SimpleStruct
    schema := s.Schema()
    
    var schemaObj map[string]interface{}
    if err := json.Unmarshal(schema, &schemaObj); err != nil {
        t.Fatal(err)
    }
    
    // Validate schema structure
    if schemaObj["type"] != "object" {
        t.Errorf("expected object type")
    }
}
```