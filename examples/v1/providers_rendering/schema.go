//go:build jsonschema

package providers_rendering

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Example) Schema() json.RawMessage { panic("not implemented") }

var _ = jsonschema.NewJSONSchemaMethod(
	Example.Schema,
	jsonschema.WithStructAccessorMethod(Example{}.A, (Example).ASchema),
	jsonschema.WithStructFunctionMethod(Example{}.B, (Example).BSchema),
	jsonschema.WithFunction(Example{}.C, BoolSchema),
	jsonschema.WithRenderProviders(), // v1: generate RenderedSchema() that executes providers
)
