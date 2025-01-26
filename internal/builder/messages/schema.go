//go:build jsonschema
// +build jsonschema

package messages

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Assertion) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (ToolFuncGetTypeInfo) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (GeneratedTestResponse) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

var (
	_ = jsonschema.NewJSONSchemaMethod(ToolFuncGetTypeInfo.Schema)
	_ = jsonschema.NewJSONSchemaMethod(GeneratedTestResponse.Schema)
	_ = jsonschema.NewInterfaceImpl[AssertionValue](
		AssertNumericValue{},
		AssertStringValue{},
		AssertBoolValue{},
		AssertType{},
		AssertArrayLength{},
	)
)
