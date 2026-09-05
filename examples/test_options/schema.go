//go:build jsonschema

package test_options

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

// Schema method stubs
func (Person) Schema() json.RawMessage   { panic("not implemented") }
func (Team) Schema() json.RawMessage     { panic("not implemented") }
func (Task) Schema() json.RawMessage     { panic("not implemented") }
func (WorkItem) Schema() json.RawMessage { panic("not implemented") }

// Test using Options pattern with a simple type
var _ = polytype.NewJSONSchemaMethod(Person.Schema)

// Team's Go doc comment supplies its schema description.
var _ = polytype.NewJSONSchemaMethod(Team.Schema)

// Register Task with its enums (Note: Severity and WeekDay will fail as global enums)
// var _ = polytype.NewJSONSchemaMethod(Task.Schema)

// Register enum types globally
var (
	// String-based enums work globally
	_ = polytype.NewEnumType[Status]()
	_ = polytype.NewEnumType[Priority]()

	// Note: Pure iota enums (Severity, WeekDay) can't be registered globally
	// They must use field-level configuration with WithEnum
)
