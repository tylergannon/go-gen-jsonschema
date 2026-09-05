//go:build jsonschema

package ptrfixture

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (*Thing) Schema() json.RawMessage { panic("not implemented") }

var _ = jsonschema.Declare((*Thing).Schema).
	Accessor(Thing{}.Name, (*Thing).NameSchema).
	Method(Thing{}.Count, (*Thing).CountSchema).
	RenderProviders()
