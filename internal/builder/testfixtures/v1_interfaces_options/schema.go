//go:build jsonschema

package v1_interfaces_options

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Owner) Schema() json.RawMessage   { panic("not implemented") }
func (Owner) ValidateJSON([]byte) error { panic("not implemented") }
func (Owner) ValidateYAML([]byte) error { panic("not implemented") }

func (Plain) Schema() json.RawMessage   { panic("not implemented") }
func (Plain) ValidateJSON([]byte) error { panic("not implemented") }
func (Plain) ValidateYAML([]byte) error { panic("not implemented") }

var _ = polytype.NewJSONSchemaMethod(Plain.Schema)

var _ = polytype.NewJSONSchemaMethod(
	Owner.Schema,
	polytype.WithInterface(
		Owner{}.IF,
		polytype.Discriminator("!kind"),
		polytype.Impl("impl_one", Impl1{}),
		polytype.Impl("impl \"two\"", Impl2{}),
	),
	polytype.WithInterface(Owner{}.IFaces),
	polytype.WithInterfaceImpls(Owner{}.IFaces, Impl1{}, Impl2{}),
	polytype.WithDiscriminator(Owner{}.IFaces, "!kind"),
	polytype.WithInterface(Owner{}.OptionalIF),
	polytype.WithInterfaceImpls(Owner{}.OptionalIF, Impl1{}, Impl2{}),
	polytype.WithDiscriminator(Owner{}.OptionalIF, "!kind"),
)
