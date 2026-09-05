//go:build jsonschema

package providers_rendering

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Example) Schema() json.RawMessage { panic("not implemented") }

// v1: RenderProviders() generates RenderedSchema() that executes providers.
var _ = polytype.Declare(Example.Schema).
	Accessor(Example{}.A, (Example).ASchema).
	Method(Example{}.B, (Example).BSchema).
	Function(Example{}.C, BoolSchema).
	RenderProviders()
