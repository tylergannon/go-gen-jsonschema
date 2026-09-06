//go:build jsonschema

package enums

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

// Schema method for Status.
// This stub will be replaced with a proper implementation during code generation.
func (Status) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for Priority.
func (Priority) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for Task.
func (Task) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for SliceOfStatus.
func (SliceOfStatus) Schema() json.RawMessage {
	panic("not implemented")
}

// These marker variables register the types with the jsonschema generator.
var (
	// Register Status for schema generation
	_ = polytype.Declare(Status.Schema)

	// Register Priority for schema generation
	_ = polytype.Declare(Priority.Schema)

	// Register Task for schema generation
	_ = polytype.Declare(Task.Schema)

	// Register SliceOfStatus for schema generation
	_ = polytype.Declare(SliceOfStatus.Schema)

	// Status and Priority need no registration here: each declares
	// `func (T) enum()` in types.go, so every use of them (Task fields and
	// the SliceOfStatus element type alike) is emitted as an enum.
)
