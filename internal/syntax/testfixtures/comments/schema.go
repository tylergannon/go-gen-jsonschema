//go:build jsonschema

package comments

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (StructType) SchemaBuilder() json.RawMessage {
	panic("not implemented")
}

var (
	_ = polytype.NewJSONSchemaMethod(StructType.SchemaBuilder)
)
