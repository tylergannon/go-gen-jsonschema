//go:build jsonschema

package interfaces_options

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Owner) Schema() json.RawMessage { panic("not implemented") }

// v1 interface options example
var _ = jsonschema.NewJSONSchemaMethod(
	Owner.Schema,
	jsonschema.WithInterface(Owner{}.IF),
	jsonschema.WithInterfaceImpls(Owner{}.IF, Impl1{}, Impl2{}),
	jsonschema.WithDiscriminator(Owner{}.IF, "!kind"),
)
