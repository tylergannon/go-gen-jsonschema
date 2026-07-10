//go:build jsonschema

package traversal

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (TraversalHolder) Schema() json.RawMessage {
	panic("not implemented")
}

var _ = jsonschema.NewJSONSchemaMethod(TraversalHolder.Schema)
