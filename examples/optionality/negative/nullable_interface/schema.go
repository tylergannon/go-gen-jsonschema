//go:build jsonschema

package nullable_interface

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Config) Schema() json.RawMessage { panic("not implemented") }

var _ = polytype.NewJSONSchemaMethod(
	Config.Schema,
	polytype.WithInterface(Config{}.Value),
	polytype.WithInterfaceImpls(Config{}.Value, Text{}),
)
