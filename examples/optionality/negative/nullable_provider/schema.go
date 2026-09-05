//go:build jsonschema

package nullable_provider

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Config) Schema() json.RawMessage { panic("not implemented") }

var _ = polytype.NewJSONSchemaMethod(
	Config.Schema,
	polytype.WithFunction(Config{}.Value, ValueSchema),
	polytype.WithRenderProviders(),
)
