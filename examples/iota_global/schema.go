//go:build jsonschema

package iota_global

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Task) Schema() json.RawMessage { panic("not implemented") }

// Priority is an iota enum that declares `func (Priority) enum()` in
// types.go; it is emitted as an integer enum with no further registration.
var _ = polytype.NewJSONSchemaMethod(Task.Schema)
