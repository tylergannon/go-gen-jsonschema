//go:build jsonschema

package sealed_interface_slices

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Batch) Schema() json.RawMessage { panic("not implemented") }

// The field selector still identifies the complete slice field; the generator
// derives the registered interface from its element type.
var _ = jsonschema.Declare(Batch.Schema).
	Interface(
		Batch{}.Events,
		jsonschema.Discriminator("!kind"),
		jsonschema.Impl("Created", Created{}),
		jsonschema.Impl("Deleted", (*Deleted)(nil)),
	)
