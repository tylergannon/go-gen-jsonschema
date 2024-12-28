package registrytestapp

import (
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

// MyInterface is a marker interface.
type MyInterface interface {
	privateIdentifier()
}

// MyStruct1 represents an entity with a coolness factor.
type MyStruct1 struct {
	// Coolness indicates how cool the entity is on a scale of 1-100.
	Coolness int `json:"coolness"`
}

func (m MyStruct1) privateIdentifier() {}

var _ MyInterface = MyStruct1{}

// MyStruct2 represents an entity with a unique name.
type MyStruct2 struct {
	// Name is the unique name of the entity.
	Name string `json:"name"`
}

func (m MyStruct2) privateIdentifier() {}

var _ MyInterface = MyStruct2{}

// MyStruct3 represents an entity with a rating and description.
type MyStruct3 struct {
	// Rating is a numerical rating of the entity.
	Rating float64 `json:"rating"`

	// Description is a short text description of the entity.
	Description string `json:"description"`
}

func (m MyStruct3) privateIdentifier() {}

var _ MyInterface = MyStruct3{}

// MyStruct4 represents an entity with dimensions.
type MyStruct4 struct {
	// Height is the height of the entity in meters.
	Height float64 `json:"height"`

	// Width is the width of the entity in meters.
	Width float64 `json:"width"`
}

func (m MyStruct4) privateIdentifier() {}

var _ MyInterface = MyStruct4{}

// MyStruct5 represents an entity with status and timestamp.
type MyStruct5 struct {
	// Status is the current status of the entity.
	Status string `json:"status"`

	// Timestamp is the last updated time for the entity.
	Timestamp int64 `json:"timestamp"`
}

func (m MyStruct5) privateIdentifier() {}

var _ MyInterface = MyStruct5{}

// MyStruct6 represents an entity with a priority level and category.
type MyStruct6 struct {
	// Priority indicates the priority level of the entity.
	Priority int `json:"priority"`

	// Category indicates the category of the entity.
	Category string `json:"category"`
}

func (m MyStruct6) privateIdentifier() {}

var _ MyInterface = MyStruct6{}

// Register all implementations of MyInterface
var _ = jsonschema.SetImplementations[MyInterface](
	MyStruct1{},
	MyStruct2{},
	MyStruct3{},
	MyStruct4{},
	MyStruct5{},
	MyStruct6{},
)
