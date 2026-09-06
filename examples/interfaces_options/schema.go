//go:build jsonschema

package interfaces_options

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Owner) Schema() json.RawMessage { panic("not implemented") }

// IFace is sealed by its unexported isIface method, so its variants (Impl1
// and Impl2) are inferred; nothing is declared at the field.
var _ = polytype.Declare(Owner.Schema)

// IFace discriminates on "!kind" instead of the default "type".
var _ = polytype.SealedUnion[IFace]("!kind")
