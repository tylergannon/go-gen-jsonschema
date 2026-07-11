//go:build jsonschema

package v1_interfaces_options

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Owner) Schema() json.RawMessage { panic("not implemented") }

var _ = jsonschema.NewJSONSchemaMethod(
	Owner.Schema,
	jsonschema.WithInterface(Owner{}.IF),
	jsonschema.WithInterfaceImpls(Owner{}.IF, Impl1{}, Impl2{}),
	jsonschema.WithDiscriminator(Owner{}.IF, "!kind"),
	jsonschema.WithInterface(Owner{}.IFaces),
	jsonschema.WithInterfaceImpls(Owner{}.IFaces, Impl1{}, Impl2{}),
	jsonschema.WithDiscriminator(Owner{}.IFaces, "!kind"),
	jsonschema.WithInterface(Owner{}.OptionalIF),
	jsonschema.WithInterfaceImpls(Owner{}.OptionalIF, Impl1{}, Impl2{}),
	jsonschema.WithDiscriminator(Owner{}.OptionalIF, "!kind"),
)
