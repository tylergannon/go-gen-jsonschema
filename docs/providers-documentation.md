# Provider Methods Documentation

## Overview

The provider methods (`WithStructAccessorMethod`, `WithStructFunctionMethod`, `WithFunction`) are part of the v1 API for go-gen-jsonschema. They enable field-level schema customization through template-based rendering.

## Purpose

These methods allow you to override the default JSON schema generation for specific struct fields by providing custom schema generation functions. When providers are configured, the generated schema becomes a template (.json.tmpl) with placeholders that get filled in at runtime.

## The Three Provider Methods

### 1. `WithStructAccessorMethod`
```go
func WithStructAccessorMethod[T, U any](val T, f func(U) json.Marshaler) SchemaMethodOption
```

**Purpose**: Uses a no-argument struct method to provide the schema for a field.

**Usage Pattern**:
```go
jsonschema.WithStructAccessorMethod(Example{}.FieldName, (Example).MethodName)
```

**When to use**: When the schema generation doesn't need access to the field's value, just the receiver.

**Example**:
```go
func (Example) ASchema() json.Marshaler {
    return json.RawMessage(`{"type":"string"}`)
}
```

### 2. `WithStructFunctionMethod`
```go
func WithStructFunctionMethod[T, U any](val U, f func(T, U) json.Marshaler) SchemaMethodOption
```

**Purpose**: Uses a struct method that takes the field value as an argument.

**Usage Pattern**:
```go
jsonschema.WithStructFunctionMethod(Example{}.FieldName, (Example).MethodName)
```

**When to use**: When the schema generation needs to examine the field's actual value.

**Example**:
```go
func (Example) BSchema(value int) json.Marshaler {
    // Can use 'value' to customize the schema
    return json.RawMessage(`{"type":"integer"}`)
}
```

### 3. `WithFunction`
```go
func WithFunction[T any](val T, f func(T) json.Marshaler) SchemaMethodOption
```

**Purpose**: Uses a free function (not a method) to provide the schema.

**Usage Pattern**:
```go
jsonschema.WithFunction(Example{}.FieldName, FreeFunction)
```

**When to use**: When the schema generation logic doesn't need receiver access.

**Example**:
```go
func BoolSchema(value bool) json.Marshaler {
    return json.RawMessage(`{"type":"boolean"}`)
}
```

## How It Works

1. **Registration**: You register providers as options to `NewJSONSchemaMethod`:
   ```go
   var _ = jsonschema.NewJSONSchemaMethod(
       Example.Schema,
       jsonschema.WithStructAccessorMethod(Example{}.A, (Example).ASchema),
       jsonschema.WithStructFunctionMethod(Example{}.B, (Example).BSchema),
       jsonschema.WithFunction(Example{}.C, BoolSchema),
       jsonschema.WithRenderProviders(), // Enable runtime rendering
   )
   ```

2. **Generation**: The tool generates:
   - A template file (`jsonschema/Example.json.tmpl`) with placeholders like `{{.a}}`, `{{.b}}`, `{{.c}}`
   - A `Schema()` method that returns the raw template
   - A `RenderedSchema()` method (if `WithRenderProviders()` is used) that executes providers and renders the template

3. **Runtime**: When `RenderedSchema()` is called:
   - Each provider function is executed with the appropriate arguments
   - Results are marshaled to JSON
   - Template is rendered with the provider results
   - Final JSON schema is returned

## Current Status

### Working ✅
- Code generation for all three provider types
- Template file generation with proper placeholders
- Provider function execution in `RenderedSchema()`
- Basic schema structure generation

### Issues Found 🔧
1. **Template Rendering Issue**: The generated `RenderedSchema()` method has a bug where provider results (json.RawMessage) are being inserted as byte arrays rather than JSON strings in the template context.

   **Symptom**: Instead of `{"type":"string"}`, the template receives `[123 34 116 121 112 101 34 58 34 115 116 114 105 110 103 34 125]`

   **Impact**: The rendered JSON is invalid because the template inserts raw byte arrays

2. **Missing Tests**: The provider examples don't have comprehensive tests to validate the rendered schemas

## Recommendations

1. **Fix Template Rendering**: The `RenderedSchema()` generation needs to be updated to properly convert json.RawMessage to string before inserting into the template context.

2. **Add Tests**: Create comprehensive tests for provider-based schema generation to ensure:
   - Templates are generated correctly
   - Providers are called with correct arguments
   - Rendered schemas are valid JSON
   - Provider values are properly substituted

3. **Documentation**: While basic documentation exists, consider adding:
   - More complex examples showing conditional schema generation
   - Best practices for provider implementation
   - Performance considerations for runtime rendering

## Example Test Case

Here's a test that demonstrates the current issue and what should work:

```go
func TestProviderRendering(t *testing.T) {
    example := Example{A: "test", B: 42, C: true}
    
    // This should work but currently produces invalid JSON
    schema, err := example.RenderedSchema()
    if err != nil {
        t.Fatal(err)
    }
    
    // Parse to verify valid JSON
    var result map[string]interface{}
    if err := json.Unmarshal(schema, &result); err != nil {
        t.Fatalf("Invalid JSON: %v\nSchema: %s", err, schema)
    }
    
    // Verify structure
    props := result["properties"].(map[string]interface{})
    assert(props["a"].(map[string]interface{})["type"] == "string")
    assert(props["b"].(map[string]interface{})["type"] == "integer")
    assert(props["c"].(map[string]interface{})["type"] == "boolean")
}
```