//go:build jsonschema

package scaffold_demo

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Widget) Schema() json.RawMessage {
	panic("not implemented")
}

var (
	_ = jsonschema.Declare(Widget.Schema)
)
