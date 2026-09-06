//go:build jsonschema

package ref_types

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Shared) Schema() json.RawMessage    { panic("not implemented") }
func (Container) Schema() json.RawMessage { panic("not implemented") }
func (NullableConfig) Schema() json.RawMessage {
	panic("not implemented")
}

// Shared is registered as its own top-level schema and, via Ref(), as a
// definition referenced from other schemas instead of being inlined there.
var _ = polytype.Declare(Shared.Schema).Ref()

var _ = polytype.Declare(Container.Schema)

var (
	_ = polytype.Declare(NullableConfig.Schema)
)
