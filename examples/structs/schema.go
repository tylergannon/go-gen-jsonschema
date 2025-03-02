//go:build jsonschema
// +build jsonschema

package structs

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

// Schema method for Address.
func (Address) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

// Schema method for ContactInfo.
func (ContactInfo) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

// Schema method for Person.
func (Person) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

// Schema method for Organization.
func (Organization) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

// Schema method for Department.
func (Department) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

// These marker variables register the types with the jsonschema generator.
var (
	// Register Address for schema generation
	_ = jsonschema.NewJSONSchemaMethod(Address.Schema)

	// Register ContactInfo for schema generation
	_ = jsonschema.NewJSONSchemaMethod(ContactInfo.Schema)

	// Register Person for schema generation
	_ = jsonschema.NewJSONSchemaMethod(Person.Schema)

	// Register Organization for schema generation
	_ = jsonschema.NewJSONSchemaMethod(Organization.Schema)

	// Register Department for schema generation
	_ = jsonschema.NewJSONSchemaMethod(Department.Schema)
)
