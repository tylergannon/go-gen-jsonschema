//go:build jsonschema
// +build jsonschema

package test_options

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

// Schema method stubs
func (Person) Schema() json.RawMessage   { panic("not implemented") }
func (Team) Schema() json.RawMessage     { panic("not implemented") }
func (Task) Schema() json.RawMessage     { panic("not implemented") }
func (WorkItem) Schema() json.RawMessage { panic("not implemented") }

// Test using Options pattern with a simple type
var _ = jsonschema.NewJSONSchemaMethod(Person.Schema)

// Test using Options pattern with more complex options
var _ = jsonschema.NewJSONSchemaMethod(
	Team.Schema,
	// Add a custom description for the Team type
	jsonschema.WithDescription("A team of people working together"),
)

// Register Task with its enums (Note: Severity and WeekDay will fail as global enums)
// var _ = jsonschema.NewJSONSchemaMethod(Task.Schema)

// Register WorkItem with field-level enum configuration (v1 pattern)
var _ = jsonschema.NewJSONSchemaMethod(
	WorkItem.Schema,
	// Configure Priority enum at field level with string mode
	jsonschema.WithEnum(WorkItem{}.Priority),
	jsonschema.WithEnumMode(jsonschema.EnumStrings),
	// Configure Severity enum at field level
	jsonschema.WithEnum(WorkItem{}.Level),
)

// Register enum types globally
var (
	// String-based enums work globally
	_ = jsonschema.NewEnumType[Status]()
	_ = jsonschema.NewEnumType[Priority]()

	// Note: Pure iota enums (Severity, WeekDay) can't be registered globally
	// They must use field-level configuration with WithEnum and WithEnumMode
)
