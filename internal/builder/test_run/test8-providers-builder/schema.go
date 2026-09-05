//go:build jsonschema

package providers_builder

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Example) Schema() json.RawMessage { panic("not implemented") }

// Register via method form with providers and rendered
var _ = polytype.NewJSONSchemaMethod(
	Example.Schema,
	polytype.WithStructAccessorMethod(Example{}.A, (Example).ASchema),
	polytype.WithStructFunctionMethod(Example{}.B, (Example).BSchema),
	polytype.WithFunction(Example{}.C, BoolSchemaFunc),
	polytype.WithRenderProviders(),
)
