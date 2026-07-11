//go:build jsonschema

package ref_types

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Shared) Schema() json.RawMessage    { panic("not implemented") }
func (Container) Schema() json.RawMessage { panic("not implemented") }

// Shared is registered as its own top-level schema and, via AsRef(), as a
// definition referenced from other schemas instead of being inlined there.
var _ = jsonschema.NewJSONSchemaMethod(Shared.Schema, jsonschema.AsRef())

var _ = jsonschema.NewJSONSchemaMethod(Container.Schema)
