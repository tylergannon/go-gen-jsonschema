//go:build jsonschema
// +build jsonschema

package structs

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (StructType1) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (StructType2) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

var (
	_ = jsonschema.NewJSONSchemaMethod(StructType1.Schema)
	_ = jsonschema.NewJSONSchemaMethod(StructType2.Schema)
	_ = jsonschema.NewEnumType[EnumType123]()
)
