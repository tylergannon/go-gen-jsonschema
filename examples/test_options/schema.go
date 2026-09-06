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

var _ = polytype.NewJSONSchemaMethod(Person.Schema)
var _ = polytype.NewJSONSchemaMethod(Team.Schema)

// Status, Priority, Severity, and WeekDay each declare `func (T) enum()` in
// enum_types.go, so Task and WorkItem need no enum registration. String
// enums emit their constant values; integer enums emit their integer values.
var _ = polytype.NewJSONSchemaMethod(Task.Schema)
var _ = polytype.NewJSONSchemaMethod(WorkItem.Schema)
