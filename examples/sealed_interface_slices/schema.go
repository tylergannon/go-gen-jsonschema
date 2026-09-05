//go:build jsonschema

package sealed_interface_slices

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Batch) Schema() json.RawMessage { panic("not implemented") }

// The field selector still identifies the complete slice field; the generator
// derives the registered interface from its element type.
var _ = polytype.Declare(Batch.Schema).
	Interface(
		Batch{}.Events,
		polytype.Discriminator("!kind"),
		polytype.Impl("Created", Created{}),
		polytype.Impl("Deleted", (*Deleted)(nil)),
	)
