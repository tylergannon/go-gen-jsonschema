//go:build jsonschema

package providers

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Example) Schema() json.RawMessage { panic("not implemented") }

// Provider implementations
func (Example) ASchema() json.Marshaler {
	return json.RawMessage(`{"type":"string","description":"A"}`)
}

func (Example) BSchema(_ int) json.Marshaler {
	return json.RawMessage(`{"type":"integer","description":"B"}`)
}

func BoolSchemaFunc(_ bool) json.Marshaler {
	return json.RawMessage(`{"type":"boolean","description":"C"}`)
}

var (
	_ = polytype.NewJSONSchemaMethod(
		Example.Schema,
		polytype.WithStructAccessorMethod(Example{}.A, (Example).ASchema),
		polytype.WithStructFunctionMethod(Example{}.B, (Example).BSchema),
		polytype.WithFunction(Example{}.C, BoolSchemaFunc),
		polytype.WithRenderProviders(),
	)
)
