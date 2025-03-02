# go-gen-jsonschema 🧩

Generate Golang-friendly JSON schemas for structured LLM responses.

<p align="center">
  <img src="gopher-front.svg" alt="Gopher mascot" width="200" height="auto">
</p>

## 🔍 Overview

`go-gen-jsonschema` automatically generates JSON Schema definitions from your Go type definitions, optimized for LLM function calling (OpenAI, Anthropic, etc). It eliminates the need to manually write and maintain JSON schemas, keeping them perfectly in sync with your Go types.

Key benefits:

- ✨ **Automatic Schema Generation**: Convert Go structs directly to JSON Schema
- 🤖 **LLM-Friendly**: Designed for AI function calling use cases
- 🛡️ **Type Safety**: Ensure LLM responses match your Go types
- 🔄 **Compile-Time Validation**: Catch schema errors during build
- 🚀 **Runtime Support**: Load schemas during execution for LLM requests

## 📦 Installation

```bash
go install github.com/tylergannon/go-gen-jsonschema/gen-jsonschema@latest
```

## 🚀 Quick Start

1. **Add type definitions** to your Go project
2. **Run the generator**:
   ```bash
   go-gen-jsonschema gen
   ```
3. **Use the generated schemas** with your LLM integration

## 💻 Command Line Usage

### 🔨 Generate Schemas

```bash
go-gen-jsonschema gen [options]
```

Options:
- `-target string`: Path to target package (defaults to current directory)
- `-pretty`: Enable pretty-printed JSON output
- `-no-gen-test`: Disable test sample generation
- `-num-test-samples int`: Number of test samples to generate (default 5)
- `-no-changes`: Fail if any schema changes are detected
- `-force`: Force regeneration of schemas even if no changes detected

### 🆕 Create a New Project

```bash
go-gen-jsonschema new [options]
```

Options:
- `-out string`: Path to output file (empty or "--" means print to stdout)
- `-pkg string`: Package name for generated file (defaults to current directory)
- `-methods string`: Comma-separated list of methods to generate (format: TypeName=MethodName,TypeName2=MethodName2)

## 📝 Examples

### 🔰 Basic Usage

1. Define your Go types:

```go
// User represents a system user
type User struct {
    ID        int    `json:"id"`
    Username  string `json:"username"`
    Email     string `json:"email"`
    CreatedAt time.Time `json:"createdAt"`
}
```

2. Run the generator:

```bash
go-gen-jsonschema gen
```

3. Use the generated schema with an LLM:

```go
schema, _ := schemas.UserSchema()
llmResponse := callLLM(prompt, schema)
var user User
json.Unmarshal(llmResponse, &user)
```

## ✨ Features

- 📝 **Doc Comment Support**: Comments become schema descriptions
- 🏷️ **JSON Tag Integration**: Respects json struct tags
- 🔒 **Type Safety**: Generates Go-compatible schemas
- 🔌 **Custom Transformers**: Extensible for special types
- ⏰ **Time Handling**: Proper formatting for time.Time
- 🧪 **Test Data Generation**: Sample data for validation

## 🛠️ Development

Build from source:

```bash
git clone https://github.com/tylergannon/go-gen-jsonschema.git
cd go-gen-jsonschema
go build ./gen-jsonschema
```

## 📄 License

[License information]

## 👥 Contributing

Contributions welcome! Please see [contributing guidelines] for more information. 
