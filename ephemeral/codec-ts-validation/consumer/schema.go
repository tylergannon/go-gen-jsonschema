//go:build jsonschema

package consumer

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Envelope) Schema() json.RawMessage { panic("not implemented") }

var _ = jsonschema.NewJSONSchemaMethod(Envelope.Schema,
	jsonschema.WithInterface(Envelope{}.Primary,
		jsonschema.Discriminator("kind"),
		jsonschema.Impl("created", Created{}),
		jsonschema.Impl("deleted", (*Deleted)(nil)),
	),
	jsonschema.WithInterface(Envelope{}.Alternate,
		jsonschema.Discriminator("!kind"),
		jsonschema.Impl("alt-created", Created{}),
		jsonschema.Impl("alt-deleted", (*Deleted)(nil)),
	),
	jsonschema.WithInterface(Envelope{}.Maybe,
		jsonschema.Discriminator("kind"),
		jsonschema.Impl("created", Created{}),
		jsonschema.Impl("deleted", (*Deleted)(nil)),
	),
	jsonschema.WithInterface(Envelope{}.Events,
		jsonschema.Discriminator("kind"),
		jsonschema.Impl("created", Created{}),
		jsonschema.Impl("deleted", (*Deleted)(nil)),
	),
	jsonschema.WithStringerEnum(Envelope{}.StringState),
	jsonschema.WithEnum(Envelope{}.NumericState),
	jsonschema.WithStringerEnum(Envelope{}.OptionalState),
	jsonschema.WithStringerEnum(Envelope{}.NullableState),
	jsonschema.WithStringerEnum(Envelope{}.NullState),
)
