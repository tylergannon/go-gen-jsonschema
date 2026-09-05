//go:build jsonschema

package inspectproof

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Supported) Schema() json.RawMessage   { panic("not implemented") }
func (Unsupported) Schema() json.RawMessage { panic("not implemented") }
func (Unknown) Schema() json.RawMessage     { panic("not implemented") }
func (OptionalMissingOmitzero) Schema() json.RawMessage {
	panic("not implemented")
}
func (AnyModel) Schema() json.RawMessage { panic("not implemented") }

var (
	_ = jsonschema.NewJSONSchemaMethod(Supported.Schema)
	_ = jsonschema.NewJSONSchemaMethod(Unsupported.Schema)
	_ = jsonschema.NewJSONSchemaMethod(Unknown.Schema)
	_ = jsonschema.NewJSONSchemaMethod(OptionalMissingOmitzero.Schema)
	_ = jsonschema.NewJSONSchemaMethod(AnyModel.Schema)
)
