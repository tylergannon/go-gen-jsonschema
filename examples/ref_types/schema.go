//go:build jsonschema

package ref_types

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Shared) Schema() json.RawMessage    { panic("not implemented") }
func (Container) Schema() json.RawMessage { panic("not implemented") }
func (NullableConfig) Schema() json.RawMessage {
	panic("not implemented")
}

// Shared is registered as its own top-level schema and, via Ref(), as a
// definition referenced from other schemas instead of being inlined there.
var _ = jsonschema.Declare(Shared.Schema).Ref()

var _ = jsonschema.Declare(Container.Schema)

var (
	_ = jsonschema.Declare(NullableConfig.Schema)
	_ = jsonschema.NewEnumType[Mode]()
)
