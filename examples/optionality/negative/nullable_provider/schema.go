//go:build jsonschema

package nullable_provider

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Config) Schema() json.RawMessage { panic("not implemented") }

var _ = jsonschema.NewJSONSchemaMethod(
	Config.Schema,
	jsonschema.WithFunction(Config{}.Value, ValueSchema),
	jsonschema.WithRenderProviders(),
)
