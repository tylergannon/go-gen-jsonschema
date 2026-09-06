//go:build jsonschema

package basictypes

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (EnumType) Schema() json.RawMessage {
	panic("not implemented")
}

func (SliceOfEnumType) Schema() json.RawMessage {
	panic("not implemented")
}

func (SliceOfRemoteEnumType) Schema() json.RawMessage {
	panic("not implemented")
}

func (SliceOfPointerToRemoteEnum) Schema() json.RawMessage {
	panic("not implemented")
}

var (
	_ = polytype.NewJSONSchemaMethod(EnumType.Schema)
	_ = polytype.NewJSONSchemaMethod(SliceOfEnumType.Schema)
	_ = polytype.NewJSONSchemaMethod(SliceOfRemoteEnumType.Schema)
	_ = polytype.NewJSONSchemaMethod(SliceOfPointerToRemoteEnum.Schema)
)
