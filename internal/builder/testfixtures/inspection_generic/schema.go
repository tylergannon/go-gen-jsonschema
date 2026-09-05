//go:build jsonschema

package inspection_generic

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Root) Schema() json.RawMessage {
	panic("not implemented")
}

var _ = jsonschema.NewJSONSchemaMethod(Root.Schema)
