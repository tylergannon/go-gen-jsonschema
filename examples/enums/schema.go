//go:build jsonschema

package enums

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
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
	_ = jsonschema.Declare(Status.Schema)

	// Register Priority for schema generation
	_ = jsonschema.Declare(Priority.Schema)

	// Register Task for schema generation
	_ = jsonschema.Declare(Task.Schema)

	// Register SliceOfStatus for schema generation
	_ = jsonschema.Declare(SliceOfStatus.Schema)

	// Kept on the legacy package-level form: Status has no fluent
	// replacement here because it's used both as a field (Task.Status) and
	// as a bare slice element type (SliceOfStatus), and field-level .Enum
	// has no chain method that applies to a slice root's own element type.
	_ = jsonschema.NewEnumType[Status]()

	// Kept alongside Status for consistency, though Priority is only used
	// as Task.Priority and would also work as .Enum(Task{}.Priority).
	_ = jsonschema.NewEnumType[Priority]()
)
