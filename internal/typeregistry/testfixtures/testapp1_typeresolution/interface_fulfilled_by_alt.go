package testapp1_typeresolution

import (
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry/testfixtures/registrytestapp/impls"
)

// MyInterface is a marker interface.
type MyInterface interface {
	SomeMethod()
}

// MyStruct1 represents an entity with a coolness factor.
type MyStruct1 struct {
	// Coolness indicates how cool the entity is on a scale of 1-100.
	Coolness int `json:"coolness"`
}

func (m MyStruct1) SomeMethod() {}

var _ MyInterface = MyStruct1{}

// MyStruct2 represents an entity with a unique name.
type MyStruct2 struct {
	// Name is the unique name of the entity.
	Name string `json:"name"`
}

func (m *MyStruct2) SomeMethod() {}

var _ MyInterface = &MyStruct2{}

// MyStruct3 represents an entity with a rating and description.
type MyStruct3 struct {
	// Rating is a numerical rating of the entity.
	Rating float64 `json:"rating"`

	// Description is a short text description of the entity.
	Description string `json:"description"`
}

func (m *MyStruct3) SomeMethod() {}

var _ MyInterface = &MyStruct3{}

// Register all implementations of MyInterface
// will result in a private function, func unmarshalMyInterfaceJSON(date []byte) (MyInterface, error)
var _ = jsonschema.SetImplementations[MyInterface](
	MyStruct1{},
	&MyStruct2{},
	(*MyStruct3)(nil),
	impls.NestedStruct1{},
	&impls.NestedStruct2{},
	(*impls.NestedStruct3)(nil),
)
