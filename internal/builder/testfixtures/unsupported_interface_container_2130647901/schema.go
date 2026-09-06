//go:build jsonschema

package fixture

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

type Variant interface{ variant() }

type First struct {
	Name string `json:"name"`
}

func (First) variant() {}

type Owner struct {
	Values polytype.Optional[[]Variant] `json:"values,omitzero"`
}

func (Owner) Schema() json.RawMessage { panic("not implemented") }

var _ = polytype.NewJSONSchemaMethod(Owner.Schema)
