//go:build jsonschema

package enums_stringmode

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Paint) Schema() json.RawMessage { panic("not implemented") }

// v1 enum string mode example
var _ = jsonschema.NewJSONSchemaMethod(
	Paint.Schema,
	jsonschema.WithEnum(Paint{}.C),
	jsonschema.WithEnumMode(jsonschema.EnumStrings),
	// Optionally: jsonschema.WithEnumName(ColorRed, "red")
)
