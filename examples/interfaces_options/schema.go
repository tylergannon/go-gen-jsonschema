//go:build jsonschema

package interfaces_options

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Owner) Schema() json.RawMessage { panic("not implemented") }

// v1 interface options example
var _ = polytype.Declare(Owner.Schema).
	Interface(
		Owner{}.IF,
		polytype.Discriminator("!kind"),
		polytype.Impl("impl_one", Impl1{}),
		polytype.Impl("impl_two", Impl2{}),
	)
