//go:build jsonschema

package enums_stringmode

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Paint) Schema() json.RawMessage { panic("not implemented") }

// Enum string mode example: Color is a numeric enum, but .StringerEnum
// renders it as a JSON string enum of its declared constant names.
var _ = jsonschema.Declare(Paint.Schema).
	StringerEnum(Paint{}.C)
