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

type Variants []Variant

func (Variants) Schema() json.RawMessage { panic("not implemented") }

var _ = polytype.NewJSONSchemaMethod(Variants.Schema)
