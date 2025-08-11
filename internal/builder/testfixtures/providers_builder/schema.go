//go:build jsonschema

package providers_builder

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func ExampleSchema() json.RawMessage { panic("not implemented") }

// Register via builder form with providers and rendered
var _ = jsonschema.NewJSONSchemaBuilderFor(
	Example{},
	ExampleSchema,
	jsonschema.WithStructAccessorMethod(Example{}.A, (Example).ASchema),
	jsonschema.WithStructFunctionMethod(Example{}.B, (Example).BSchema),
	jsonschema.WithFunction(Example{}.C, BoolSchemaFunc),
	jsonschema.WithRenderProviders(),
)
