//go:build jsonschema

package consumer

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Envelope) Schema() json.RawMessage    { panic("not implemented") }
func (Detail) Schema() json.RawMessage      { panic("not implemented") }
func (Composition) Schema() json.RawMessage { panic("not implemented") }

var _ = jsonschema.NewJSONSchemaMethod(Detail.Schema, jsonschema.AsRef())
var _ = jsonschema.NewJSONSchemaMethod(Composition.Schema)
var _ = jsonschema.NewEnumType[Status]()
var _ = jsonschema.NewJSONSchemaMethod(Envelope.Schema,
	jsonschema.WithInterface(Envelope{}.Event,
		jsonschema.Discriminator("!kind"),
		jsonschema.Impl("created", Created{}),
		jsonschema.Impl("deleted", (*Deleted)(nil)),
		// ADDED_VARIANT
	),
	jsonschema.WithInterface(Envelope{}.Other,
		jsonschema.Discriminator("other-key"),
		jsonschema.Impl("create\"雪", Created{}),
	),
	jsonschema.WithInterface(Envelope{}.Maybe,
		jsonschema.Discriminator("!kind"),
		jsonschema.Impl("created", Created{}),
		jsonschema.Impl("deleted", (*Deleted)(nil)),
	),
	jsonschema.WithInterface(Envelope{}.Events,
		jsonschema.Discriminator("!kind"),
		jsonschema.Impl("created", Created{}),
		jsonschema.Impl("deleted", (*Deleted)(nil)),
	),
	jsonschema.WithEnum(Envelope{}.Priority),
	jsonschema.WithStringerEnum(Envelope{}.PriorityName),
)
