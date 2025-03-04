//go:build jsonschema
// +build jsonschema

package comments

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (StructType) SchemaBuilder() json.RawMessage {
	panic("not implemented")
}

var (
	_ = jsonschema.NewEnumType[StringType]()
	_ = jsonschema.NewJSONSchemaMethod(StructType.SchemaBuilder)
)
