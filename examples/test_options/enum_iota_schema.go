//go:build jsonschema
// +build jsonschema

package test_options

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

// Schema method stubs for iota enum types
func (Product) Schema() json.RawMessage       { panic("not implemented") }
func (Configuration) Schema() json.RawMessage { panic("not implemented") }

// Register Product - iota enums will be represented as integers by default
// var _ = jsonschema.NewJSONSchemaMethod(Product.Schema)
// This would crash because iota enums can't be used without field-level config

// COMMENTED OUT: EnumContainer type not defined
// // Register Configuration with field-level enum configuration to use string mode
// var _ = jsonschema.NewJSONSchemaMethod(
// 	EnumContainer.Schema,
// 	jsonschema.WithEnum(EnumContainer{}.MySimpleEnum),
// )

// IMPORTANT: Pure iota-based enums CANNOT be registered globally with NewEnumType
// They must use field-level configuration with WithEnum
// Only string-based enums can use NewEnumType globally
