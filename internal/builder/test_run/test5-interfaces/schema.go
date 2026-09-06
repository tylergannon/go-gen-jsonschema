//go:build jsonschema

package interfaces

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (FancyStruct) Schema() json.RawMessage {
	panic("not implemented")
}

var (
	// identifies FancyStruct as a type that should be given a schema, and
	// the `Schema()` struct method as the one that should be wired to provide
	// the generated JSON schema.
	_ = polytype.NewJSONSchemaMethod(FancyStruct.Schema)
	// Identifies TestInterface as a marked interface having known
	// implementations.  In this case there are three implementations of the
	// TestInterface interface, which will go in to the union type.
	_ = polytype.NewInterfaceImpl[TestInterface](TestInterface1{}, TestInterface2{}, (*PointerToTestInterface)(nil))
)
