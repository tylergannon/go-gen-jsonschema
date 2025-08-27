//go:build jsonschema

package enums_stringmode

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Paint) Schema() json.RawMessage { panic("not implemented") }

// v1 enum string mode example
var _ = jsonschema.NewJSONSchemaMethod(
	TestCase.Schema,
	jsonschema.WithEnum(TestCase{}.Priority),
)
