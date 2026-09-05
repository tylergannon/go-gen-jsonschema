//go:build jsonschema

package optional_without_omitzero

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Config) Schema() json.RawMessage { panic("not implemented") }

var _ = polytype.NewJSONSchemaMethod(Config.Schema)
