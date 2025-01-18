package structs

import (
	_ "github.com/dave/dst"
	_ "github.com/tylergannon/structtag"
)

//go:generate go run github.com/tylergannon/go-gen-jsonschema/gen-jsonschema/ --pretty

// It's really that way
type EnumType123 string

const (
	Var1 EnumType123 = "var1"
	Var2 EnumType123 = "var2"
	Var3 EnumType123 = "var3"
	Var4 EnumType123 = "var4"
)

// Here's what the StructType1 is all about
type StructType1 struct {
	// Foobar brain baz bag
	Field1 string `json:"field1"`
	// Test field 2
	Field2 string `json:"field2"`
	// More comments
	Field3 string `json:"field3"`
	// Tell me about zoos and things
	Field4 string `json:"field4"`

	Field5 [][]struct {
		// The second field is truly interesting.
		Field2 []EnumType123 `json:"field2"`
		Field3 struct {
			Field9 []*EnumType123 `json:"field9"`
			// foobar is just a field where you do things
			Foobar string `json:"foobar"`
		}
	} `json:"field5"`
}

// Tell me a story about fairy tails and other things
type StructType2 struct {
	StructType1

	// NestedStruct is very interesting
	NestedStruct []StructType1 `json:"nestedStruct"`
}
