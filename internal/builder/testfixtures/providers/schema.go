//go:build jsonschema
// +build jsonschema

package providers

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
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
	_ = jsonschema.NewJSONSchemaMethod(
		Example.Schema,
		jsonschema.WithStructAccessorMethod(Example{}.A, (Example).ASchema),
		jsonschema.WithStructFunctionMethod(Example{}.B, (Example).BSchema),
		jsonschema.WithFunction(Example{}.C, BoolSchemaFunc),
	)
)
