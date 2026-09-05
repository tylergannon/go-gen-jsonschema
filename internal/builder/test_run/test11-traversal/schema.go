//go:build jsonschema

package traversal

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (TraversalHolder) Schema() json.RawMessage {
	panic("not implemented")
}

var _ = polytype.NewJSONSchemaMethod(TraversalHolder.Schema)
