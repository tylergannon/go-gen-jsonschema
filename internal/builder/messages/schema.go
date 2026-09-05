//go:build jsonschema

package messages

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Assertion) Schema() json.RawMessage {
	panic("not implemented")
}

func (ToolFuncGetTypeInfo) Schema() json.RawMessage {
	panic("not implemented")
}

func (GeneratedTestResponse) Schema() json.RawMessage {
	panic("not implemented")
}

var (
	_ = polytype.NewJSONSchemaMethod(ToolFuncGetTypeInfo.Schema)
	_ = polytype.NewJSONSchemaMethod(GeneratedTestResponse.Schema)
	_ = polytype.NewInterfaceImpl[AssertionValue](
		AssertNumericValue{},
		AssertStringValue{},
		AssertBoolValue{},
		AssertType{},
		AssertArrayLength{},
	)
)
