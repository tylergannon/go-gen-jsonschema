//go:build jsonschema

package structs

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

// Schema method for Address.
func (Address) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for ContactInfo.
func (ContactInfo) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for RetryPolicy.
func (RetryPolicy) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for Person.
func (Person) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for Organization.
func (Organization) Schema() json.RawMessage {
	panic("not implemented")
}

// Schema method for Department.
func (Department) Schema() json.RawMessage {
	panic("not implemented")
}

// These marker variables register the types with the jsonschema generator.
var (
	// Register Address for schema generation
	_ = jsonschema.Declare(Address.Schema)

	// Register ContactInfo for schema generation
	_ = jsonschema.Declare(ContactInfo.Schema)

	// Register RetryPolicy for schema generation
	_ = jsonschema.Declare(RetryPolicy.Schema)

	// Register Person for schema generation
	_ = jsonschema.Declare(Person.Schema)

	// Register Organization for schema generation
	_ = jsonschema.Declare(Organization.Schema)

	// Register Department for schema generation
	_ = jsonschema.Declare(Department.Schema)
)
