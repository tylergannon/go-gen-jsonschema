//go:build jsonschema
// +build jsonschema

package basictypes

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

// Schema method for SimpleInt.
// This stub will be replaced with a proper implementation during code generation.
// The method signature must match exactly: json.RawMessage
func (SimpleInt) Schema() json.RawMessage {
	panic("not implemented") // This will be replaced by the generator
}

// Schema method for SimpleString.
func (SimpleString) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for SimpleFloat.
func (SimpleFloat) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for SimpleStruct.
func (SimpleStruct) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for TypeInNestedDecl.
func (TypeInNestedDecl) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for AnotherNestedType.
func (AnotherNestedType) Schema() json.RawMessage {
	panic("not implemented")
}

// These marker variables register the types with the jsonschema generator.
// Each type that needs a schema must be registered here using NewJSONSchemaMethod.
var (
	// Register SimpleInt for schema generation
	_ = jsonschema.NewJSONSchemaMethod(SimpleInt.Schema)

	// Register SimpleString for schema generation
	_ = jsonschema.NewJSONSchemaMethod(SimpleString.Schema)

	// Register SimpleFloat for schema generation
	_ = jsonschema.NewJSONSchemaMethod(SimpleFloat.Schema)

	// Register SimpleStruct for schema generation
	_ = jsonschema.NewJSONSchemaMethod(SimpleStruct.Schema)

	// Register TypeInNestedDecl for schema generation
	_ = jsonschema.NewJSONSchemaMethod(TypeInNestedDecl.Schema)

	// Register AnotherNestedType for schema generation
	_ = jsonschema.NewJSONSchemaMethod(AnotherNestedType.Schema)
)
