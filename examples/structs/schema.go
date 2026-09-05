//go:build jsonschema

package structs

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
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
	_ = polytype.Declare(Address.Schema)

	// Register ContactInfo for schema generation
	_ = polytype.Declare(ContactInfo.Schema)

	// Register RetryPolicy for schema generation
	_ = polytype.Declare(RetryPolicy.Schema)

	// Register Person for schema generation
	_ = polytype.Declare(Person.Schema)

	// Register Organization for schema generation
	_ = polytype.Declare(Organization.Schema)

	// Register Department for schema generation
	_ = polytype.Declare(Department.Schema)
)
