//go:build jsonschema

package stringer_enums

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

// ApplicationConfig schema with WithStringerEnum (WITHOUT NewEnumType - testing auto-discovery!)
func (ApplicationConfig) Schema() json.RawMessage { panic("not implemented") }

var _ = polytype.Declare(ApplicationConfig.Schema).
	StringerEnum(ApplicationConfig{}.LogLevel).
	StringerEnum(ApplicationConfig{}.DefaultPriority)

// Task schema with regular Enum (also WITHOUT NewEnumType!)
func (Task) Schema() json.RawMessage { panic("not implemented") }

var _ = polytype.Declare(Task.Schema).
	Enum(Task{}.Priority). // This will use integer values
	Enum(Task{}.LogLevel)  // This will use integer values
