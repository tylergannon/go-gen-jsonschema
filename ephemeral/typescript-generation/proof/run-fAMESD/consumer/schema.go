//go:build jsonschema

package consumer

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Envelope) Schema() json.RawMessage    { panic("not implemented") }
func (Detail) Schema() json.RawMessage      { panic("not implemented") }
func (Composition) Schema() json.RawMessage { panic("not implemented") }

var _ = polytype.NewJSONSchemaMethod(Detail.Schema, polytype.AsRef())
var _ = polytype.NewJSONSchemaMethod(Composition.Schema)
var _ = polytype.NewJSONSchemaMethod(Envelope.Schema,
	polytype.WithInterface(Envelope{}.Event,
		polytype.Discriminator("!kind"),
		polytype.Impl("created", Created{}),
		polytype.Impl("deleted", (*Deleted)(nil)),
		// ADDED_VARIANT
	),
	polytype.WithInterface(Envelope{}.Other,
		polytype.Discriminator("other-key"),
		polytype.Impl("create\"雪", Created{}),
	),
	polytype.WithInterface(Envelope{}.Maybe,
		polytype.Discriminator("!kind"),
		polytype.Impl("created", Created{}),
		polytype.Impl("deleted", (*Deleted)(nil)),
	),
	polytype.WithInterface(Envelope{}.Events,
		polytype.Discriminator("!kind"),
		polytype.Impl("created", Created{}),
		polytype.Impl("deleted", (*Deleted)(nil)),
	),
	polytype.WithStringerEnum(Envelope{}.PriorityName),
)
