//go:build jsonschema

package enums_stringmode

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Paint) Schema() json.RawMessage { panic("not implemented") }

// v1 enum string mode example: Color is a numeric enum, but WithStringerEnum
// renders it as a JSON string enum of its declared constant names.
var _ = jsonschema.NewJSONSchemaMethod(
	Paint.Schema,
	jsonschema.WithStringerEnum(Paint{}.C),
)
