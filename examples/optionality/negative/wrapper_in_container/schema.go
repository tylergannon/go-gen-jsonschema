//go:build jsonschema

package wrapper_in_container

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Config) Schema() json.RawMessage { panic("not implemented") }

var _ = polytype.NewJSONSchemaMethod(Config.Schema)
