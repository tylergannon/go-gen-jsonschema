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

## 🚀 Quickstart

This quickstart guide will walk you through setting up go-gen-jsonschema and implementing various schema types including basic types, enums, and union types via interfaces.

### 1. Set up your schema.go file

Create a `schema.go` file in your package with the following build tags and imports:

```go
//go:build jsonschema
// +build jsonschema

package yourpackage

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)
```

### 2. Define schema methods for your types

For each type that needs a JSON schema, add a `Schema()` method stub:

```go
func (YourType) Schema() (json.RawMessage, error) {
	panic("not implemented") // This will be replaced by the generator
}
```

### 3. Register your types with marker functions

Use the marker functions to register your types for schema generation:

```go
var (
	// Register schema methods for your types
	_ = jsonschema.NewJSONSchemaMethod(YourType.Schema)
	
	// For enums (string-based only for now)
	_ = jsonschema.NewEnumType[EnumType]()
	
	// For union types via interfaces
	_ = jsonschema.NewInterfaceImpl[YourInterface](Implementation1{}, Implementation2{}, (*PointerImplementation)(nil))
)
```

### 4. Create a generator

Create a folder named `gen` with a main.go file:

```go
package main

import (
	"log"
	
	"github.com/tylergannon/go-gen-jsonschema/gen-jsonschema/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
```

### 5. Add a go:generate directive

In your types.go file, add:

```go
//go:generate go run ./gen
```

### 6. Generate your schemas

Run:

```bash
go generate ./...
```

The generator will create JSON schema files in a `jsonschema` directory and a `jsonschema_gen.go` file with functions to access these schemas at runtime.

## 🔧 Marker Functions Explained

`go-gen-jsonschema` uses marker functions to identify and configure types for schema generation.

### NewJSONSchemaMethod

```go
_ = jsonschema.NewJSONSchemaMethod(YourType.Schema)
```

This marker registers a struct method as a stub that will be implemented with a proper JSON schema. Use this for all types that need schemas.

### NewEnumType

```go
_ = jsonschema.NewEnumType[EnumType]()
```

Marks a type as an enum. This will generate an enum schema with all const values of this type defined in the same package. Currently only string-based enums are supported.

### NewInterfaceImpl

```go
_ = jsonschema.NewInterfaceImpl[YourInterface](Implementation1{}, Implementation2{}, (*PointerImplementation)(nil))
```

Creates a union type from an interface. Pass all implementations of the interface as arguments. For pointer receivers, use `(*Type)(nil)` syntax.

### NewJSONSchemaBuilder

```go
_ = jsonschema.NewJSONSchemaBuilder[YourType](SchemaFunction)
```

Similar to NewJSONSchemaMethod but for standalone functions rather than struct methods.

## 📋 Examples

### 🔰 Basic Types

```go
// types.go
package example

type UserID int
type Username string

type User struct {
    ID       UserID   `json:"id"`
    Username Username `json:"username"`
    Email    string   `json:"email"`
}

// schema.go
//go:build jsonschema
// +build jsonschema

package example

import (
    "encoding/json"
    jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (User) Schema() (json.RawMessage, error) {
    panic("not implemented")
}

var (
    _ = jsonschema.NewJSONSchemaMethod(User.Schema)
)
```

### 🎯 Enum Types

```go
// types.go
package example

type Role string

const (
    RoleAdmin    Role = "admin"
    RoleUser     Role = "user"
    RoleGuest    Role = "guest"
)

type UserWithRole struct {
    Username string `json:"username"`
    Role     Role   `json:"role"`
}

// schema.go
//go:build jsonschema
// +build jsonschema

package example

import (
    "encoding/json"
    jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (UserWithRole) Schema() (json.RawMessage, error) {
    panic("not implemented")
}

var (
    _ = jsonschema.NewJSONSchemaMethod(UserWithRole.Schema)
    _ = jsonschema.NewEnumType[Role]()
)
```

### 🔄 Union Types via Interfaces

```go
// types.go
package example

type PaymentMethod interface {
    IsPaymentMethod()
}

type CreditCard struct {
    CardNumber string `json:"cardNumber"`
    Expiry     string `json:"expiry"`
    CVV        string `json:"cvv"`
}

func (CreditCard) IsPaymentMethod() {}

type BankTransfer struct {
    AccountNumber string `json:"accountNumber"`
    RoutingNumber string `json:"routingNumber"`
}

func (BankTransfer) IsPaymentMethod() {}

type PayPal struct {
    Email string `json:"email"`
}

func (*PayPal) IsPaymentMethod() {}

type Payment struct {
    Amount        float64       `json:"amount"`
    PaymentMethod PaymentMethod `json:"paymentMethod"`
}

// schema.go
//go:build jsonschema
// +build jsonschema

package example

import (
    "encoding/json"
    jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Payment) Schema() (json.RawMessage, error) {
    panic("not implemented")
}

var (
    _ = jsonschema.NewJSONSchemaMethod(Payment.Schema)
    _ = jsonschema.NewInterfaceImpl[PaymentMethod](CreditCard{}, BankTransfer{}, (*PayPal)(nil))
)
```

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
