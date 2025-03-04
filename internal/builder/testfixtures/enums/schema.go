//go:build jsonschema
// +build jsonschema

package basictypes

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
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
	_ = jsonschema.NewJSONSchemaMethod(EnumType.Schema)
	_ = jsonschema.NewJSONSchemaMethod(SliceOfEnumType.Schema)
	_ = jsonschema.NewJSONSchemaMethod(SliceOfRemoteEnumType.Schema)
	_ = jsonschema.NewJSONSchemaMethod(SliceOfPointerToRemoteEnum.Schema)
	_ = jsonschema.NewEnumType[EnumType]()
)
