//go:build jsonschema

package fixture

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

type Example struct {
	A string `json:"a"`
	B int    `json:"b"`
	C bool   `json:"c"`
}

func (Example) Schema() json.RawMessage { panic("not implemented") }
func (Example) ASchema() json.Marshaler {
	return json.RawMessage(`{"type":"string","description":"A"}`)
}
func (Example) BSchema(_ int) json.Marshaler {
	return json.RawMessage(`{"type":"integer","description":"B"}`)
}
func BoolSchemaFunc(_ bool) json.Marshaler {
	return json.RawMessage(`{"type":"boolean","description":"C"}`)
}

var _ = polytype.NewJSONSchemaMethod(
	Example.Schema,
	polytype.WithStructAccessorMethod(Example{}.A, (Example).ASchema),
	polytype.WithStructFunctionMethod(Example{}.B, (Example).BSchema),
	polytype.WithFunction(Example{}.C, BoolSchemaFunc),
	polytype.WithRenderProviders(),
)
