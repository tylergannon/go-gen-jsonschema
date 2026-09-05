package test_options

import "github.com/tylergannon/polytype"

//go:generate go run ../../polytype/

// Person represents a person with basic information
type Person struct {
	// Name is the person's full name
	Name string `json:"name"`

	// Age is the person's age in years
	Age int `json:"age"`

	// Email is an optional email address
	Email polytype.Optional[string] `json:"email,omitzero"`
}

// Team represents a team of people
type Team struct {
	// TeamName is the name of the team
	TeamName string `json:"teamName"`

	// Members is a list of people in the team
	Members []Person `json:"members"`
}
