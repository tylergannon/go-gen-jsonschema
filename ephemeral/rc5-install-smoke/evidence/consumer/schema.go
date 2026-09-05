//go:build jsonschema

package adoptiontransport

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Shipment) Schema() json.RawMessage     { panic("not implemented") }
func (Shipment) ValidateJSON(_ []byte) error { panic("not implemented") }

var _ = jsonschema.NewJSONSchemaMethod(Shipment.Schema,
	jsonschema.WithStringerEnum(Shipment{}.Status),
	jsonschema.WithInterface(Shipment{}.Event,
		jsonschema.Discriminator("kind"),
		jsonschema.Impl("created", Created{}),
		jsonschema.Impl("dispatched", Dispatched{}),
	),
)
