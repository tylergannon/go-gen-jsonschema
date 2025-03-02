//go:build jsonschema
// +build jsonschema

package enums

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

// Schema method for Status.
// This stub will be replaced with a proper implementation during code generation.
func (Status) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

// Schema method for Priority.
func (Priority) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

// Schema method for Task.
func (Task) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

// Schema method for SliceOfStatus.
func (SliceOfStatus) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

// These marker variables register the types with the jsonschema generator.
var (
	// Register Status for schema generation
	_ = jsonschema.NewJSONSchemaMethod(Status.Schema)

	// Register Priority for schema generation
	_ = jsonschema.NewJSONSchemaMethod(Priority.Schema)

	// Register Task for schema generation
	_ = jsonschema.NewJSONSchemaMethod(Task.Schema)

	// Register SliceOfStatus for schema generation
	_ = jsonschema.NewJSONSchemaMethod(SliceOfStatus.Schema)

	// Mark Status as an enum type
	// This tells the generator to treat Status as an enum and include all defined
	// constant values of this type in the schema.
	_ = jsonschema.NewEnumType[Status]()

	// Mark Priority as an enum type
	_ = jsonschema.NewEnumType[Priority]()
)
