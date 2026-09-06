//go:build jsonschema

package iota_global

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Task) Schema() json.RawMessage { panic("not implemented") }

// Priority is a pure iota int enum. Iota enums can't be registered with the
// global NewEnumType[T ~string](); they need field-level enum registration
// instead. Trade-off: field-level registration doesn't pick up Priority's
// own doc comment as a description (unlike a globally-registered enum
// type), so "priority" below has no "description" in the generated schema.
var _ = polytype.Declare(Task.Schema).
	Enum(Task{}.Priority)
