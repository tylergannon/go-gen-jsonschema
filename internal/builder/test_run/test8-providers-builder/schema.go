//go:build jsonschema

package providers_builder

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Example) Schema() json.RawMessage { panic("not implemented") }

// Register via method form with providers and rendered
var _ = jsonschema.NewJSONSchemaMethod(
	Example.Schema,
	jsonschema.WithStructAccessorMethod(Example{}.A, (Example).ASchema),
	jsonschema.WithStructFunctionMethod(Example{}.B, (Example).BSchema),
	jsonschema.WithFunction(Example{}.C, BoolSchemaFunc),
	jsonschema.WithRenderProviders(),
)
