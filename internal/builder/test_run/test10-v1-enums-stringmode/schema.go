//go:build jsonschema

package v1_enums_stringmode

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Paint) Schema() json.RawMessage { panic("not implemented") }

var _ = jsonschema.NewJSONSchemaMethod(
	Paint.Schema,
	jsonschema.WithEnum(Paint{}.C),
	jsonschema.WithEnumMode(jsonschema.EnumStrings),
)
