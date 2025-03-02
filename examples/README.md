# go-gen-jsonschema Examples

This directory contains complete, working examples of how to use go-gen-jsonschema to generate JSON Schema definitions from Go types. Each subdirectory demonstrates a different feature or pattern.

## Getting Started

Each example directory contains two main files:

1. `types.go` - Contains the Go type definitions with documentation
2. `schema.go` - Contains the schema stubs and marker functions

To generate the schemas for any example, run:

```bash
cd [example-directory]
go generate ./...
```

This will create a `jsonschema` directory with the generated JSON Schema files and a `jsonschema_gen.go` file with Go code for accessing the schemas at runtime.

## Example Directories

### 1. basictypes

Demonstrates how to generate schemas for basic Go types like integers, strings, and simple structs. Shows how to:

- Define basic types with documentation
- Register types for schema generation
- Handle nested type declarations

### 2. enums

Demonstrates how to create and register enum types in Go. Shows how to:

- Define string-based enum types with constant values
- Register enums with the `NewEnumType` marker function
- Use enums within other structures
- Handle slices of enum types

### 3. structs

Demonstrates complex struct types with various field types and relationships. Shows how to:

- Create deeply nested struct hierarchies
- Use embedded structs
- Handle complex field types like maps and slices
- Use time.Time fields
- Create recursive struct types (e.g., tree structures)

### 4. indirecttypes

Demonstrates various forms of type indirection. Shows how to:

- Define and use pointer types
- Work with slices of basic and custom types
- Handle slices of pointers
- Create named types based on other types
- Use maps with various value types

### 5. uniontypes

Demonstrates how to create union types using interfaces. Shows how to:

- Define interfaces that can be implemented by multiple types
- Register interface implementations for union types
- Work with both value and pointer receivers
- Use union types within structs
- Create discriminated unions with the !type field

## Key Concepts

### 1. Build Tags

All `schema.go` files use build tags to ensure they're only compiled during schema generation:

```go
//go:build jsonschema
// +build jsonschema
```

### 2. Schema Method Stubs

Each type that needs a schema requires a method stub:

```go
func (YourType) Schema() (json.RawMessage, error) {
    panic("not implemented")
}
```

### 3. Marker Variables

Types are registered using marker variables in a var block:

```go
var (
    _ = jsonschema.NewJSONSchemaMethod(YourType.Schema)
    _ = jsonschema.NewEnumType[EnumType]()
    _ = jsonschema.NewInterfaceImpl[YourInterface](Impl1{}, Impl2{}, (*PtrImpl)(nil))
)
```

### 4. Generated Files

After running `go generate`, you'll see:

- `jsonschema/` directory with JSON Schema files (one per type)
- `jsonschema_gen.go` with functions to access schemas at runtime 