//go:build jsonschema

package union_codec

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Envelope) Schema() json.RawMessage   { panic("not implemented") }
func (Envelope) ValidateJSON([]byte) error { panic("not implemented") }
func (Nested) Schema() json.RawMessage     { panic("not implemented") }

var _ = polytype.NewJSONSchemaMethod(
	Envelope.Schema,
	polytype.WithInterface(
		Envelope{}.Primary,
		polytype.Discriminator("!kind"),
		polytype.Impl("created", Created{}),
		polytype.Impl("deleted", (*Deleted)(nil)),
		polytype.Impl("", Empty{}),
	),
	polytype.WithInterface(Envelope{}.Events),
	polytype.WithInterfaceImpls(Envelope{}.Events, Created{}, (*Deleted)(nil)),
	polytype.WithInterface(Envelope{}.Optional),
	polytype.WithInterfaceImpls(Envelope{}.Optional, Created{}, (*Deleted)(nil)),
	polytype.WithInterface(
		Envelope{}.Alternate,
		polytype.Discriminator("kind\"quoted"),
		polytype.Impl("new\"event", Created{}),
		polytype.Impl("gone", (*Deleted)(nil)),
	),
	polytype.WithInterface(
		Envelope{}.Single,
		polytype.Discriminator("single"),
		polytype.Impl("only", Created{}),
	),
	polytype.WithInterface(
		Envelope{}.Hook,
		polytype.Discriminator("hookKind"),
		polytype.Impl("hooked", Hooked{}),
	),
	polytype.WithInterface(
		Envelope{}.ValueHook,
		polytype.Discriminator("valueHookKind"),
		polytype.Impl("value-hook", PointerHookValue{}),
	),
	polytype.WithStringerEnum(Envelope{}.State),
)

var _ = polytype.NewJSONSchemaMethod(
	Nested.Schema,
	polytype.WithInterface(
		Nested{}.Event,
		polytype.Discriminator("nestedKind"),
		polytype.Impl("nested-created", Created{}),
		polytype.Impl("nested-deleted", (*Deleted)(nil)),
	),
)
