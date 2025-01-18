package interfaces

import (
	_ "github.com/dave/dst"
	_ "github.com/tylergannon/structtag"
)

//go:generate go run github.com/tylergannon/go-gen-jsonschema/gen-jsonschema/ --pretty

type TestInterface interface {
	marker()
}

// Make this look pretty interesting.
type FancyStruct struct {
	// Here are the details.  Make sure you fill them out.
	Details [](*struct {
		Name      string `json:"-"`
		OtherName string `json:"-"`
		funk      int
		// Highly interesting stuff regarding Foo and Bar.
		Foo, Bar string
		// Something tells me this isn't going to make it into the document.
		IFace TestInterface `json:"iface"`
	})
}

// Put this down when you feel really great about life.
type TestInterface1 struct {
	Field1 string `json:"field1"` // obvious
	Field2 string `json:"field2"` // oblivious
	Field3 int    `json:"field3"` // obsequious
}

func (t TestInterface1) marker() {}

// This is seriously silly, don't you imagine so?
type TestInterface2 struct {
	Fork3 int `json:"fork3"`
	Fork4 int `json:"fork4"`
	Fork5 int `json:"fork5"`
}

func (t TestInterface2) marker() {}
