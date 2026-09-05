//go:build jsonschema

package iota_global

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Task) Schema() json.RawMessage { panic("not implemented") }

// Priority is a pure iota int enum. Iota enums can't be registered with the
// global NewEnumType[T ~string](); they need field-level enum registration.
var _ = polytype.Declare(Task.Schema).
	Enum(Task{}.Priority)
