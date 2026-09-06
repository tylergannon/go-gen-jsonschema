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
	// TestInterface is sealed by its unexported marker method, so its three
	// implementations (two value variants and one pointer variant) are
	// inferred and go into the union type without any declaration here.
)
