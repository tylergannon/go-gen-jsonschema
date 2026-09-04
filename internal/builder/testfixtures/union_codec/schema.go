//go:build jsonschema

package union_codec

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Envelope) Schema() json.RawMessage   { panic("not implemented") }
func (Envelope) ValidateJSON([]byte) error { panic("not implemented") }
func (Nested) Schema() json.RawMessage     { panic("not implemented") }

var _ = jsonschema.NewJSONSchemaMethod(
	Envelope.Schema,
	jsonschema.WithInterface(
		Envelope{}.Primary,
		jsonschema.Discriminator("!kind"),
		jsonschema.Impl("created", Created{}),
		jsonschema.Impl("deleted", (*Deleted)(nil)),
		jsonschema.Impl("", Empty{}),
	),
	jsonschema.WithInterface(Envelope{}.Events),
	jsonschema.WithInterfaceImpls(Envelope{}.Events, Created{}, (*Deleted)(nil)),
	jsonschema.WithInterface(Envelope{}.Optional),
	jsonschema.WithInterfaceImpls(Envelope{}.Optional, Created{}, (*Deleted)(nil)),
	jsonschema.WithInterface(
		Envelope{}.Alternate,
		jsonschema.Discriminator("kind\"quoted"),
		jsonschema.Impl("new\"event", Created{}),
		jsonschema.Impl("gone", (*Deleted)(nil)),
	),
	jsonschema.WithInterface(
		Envelope{}.Single,
		jsonschema.Discriminator("single"),
		jsonschema.Impl("only", Created{}),
	),
	jsonschema.WithInterface(
		Envelope{}.Hook,
		jsonschema.Discriminator("hookKind"),
		jsonschema.Impl("hooked", Hooked{}),
	),
)

var _ = jsonschema.NewJSONSchemaMethod(
	Nested.Schema,
	jsonschema.WithInterface(
		Nested{}.Event,
		jsonschema.Discriminator("nestedKind"),
		jsonschema.Impl("nested-created", Created{}),
		jsonschema.Impl("nested-deleted", (*Deleted)(nil)),
	),
)
