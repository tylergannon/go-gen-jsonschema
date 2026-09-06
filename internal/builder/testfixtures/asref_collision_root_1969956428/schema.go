//go:build jsonschema

package fixture

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
	dep "github.com/tylergannon/polytype/internal/builder/testfixtures/asref_collision_dep_1423005799"
)

type Shared struct {
	Name string `json:"name"`
}

func (Shared) Schema() json.RawMessage { panic("not implemented") }

type Container struct {
	Local  Shared     `json:"local"`
	Remote dep.Shared `json:"remote"`
}

func (Container) Schema() json.RawMessage { panic("not implemented") }

var (
	_ = polytype.NewJSONSchemaMethod(Shared.Schema, polytype.AsRef())
	_ = polytype.NewJSONSchemaMethod(dep.Shared.Schema, polytype.AsRef())
	_ = polytype.NewJSONSchemaMethod(Container.Schema)
)
