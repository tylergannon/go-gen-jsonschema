//go:build jsonschema

package sealed_interface_slices

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Batch) Schema() json.RawMessage { panic("not implemented") }

// The field selector still identifies the complete slice field; the generator
// derives the registered interface from its element type.
var _ = jsonschema.NewJSONSchemaMethod(
	Batch.Schema,
	jsonschema.WithInterface(Batch{}.Events),
	jsonschema.WithInterfaceImpls(Batch{}.Events, Created{}, (*Deleted)(nil)),
	jsonschema.WithDiscriminator(Batch{}.Events, "!kind"),
)
