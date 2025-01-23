//go:build jsonschema
// +build jsonschema

package interfaces

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (FancyStruct) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

var (
	// identifies FancyStruct as a type that should be given a schema, and
	// the `Schema()` struct method as the one that should be wired to provide
	// the generated JSON schema.
	_ = jsonschema.NewJSONSchemaMethod(FancyStruct.Schema)
	// Identifies TestInterface as a marked interface having known
	// implementations.  In this case there are three implementations of the
	// TestInterface interface, which will go in to the union type.
	_ = jsonschema.NewInterfaceImpl[TestInterface](TestInterface1{}, TestInterface2{}, (*PointerToTestInterface)(nil))
	// Identifies MyEnumType as an enum.  Instances of MyEnumType will
	// therefore be described as enums in the schema.  All `const` values of
	// this type defined within the same package will become possible values.
	_ = jsonschema.NewEnumType[MyEnumType]()
)
