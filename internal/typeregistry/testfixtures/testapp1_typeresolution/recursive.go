package testapp1_typeresolution

import jsonschema "github.com/tylergannon/go-gen-jsonschema"

type ParentStruct struct {
	Inline struct {
		Bar    *int
		Baz    *string
		Coolio *bool
		Child  ChildStruct
		//Children []ChildStruct
		//Nieces   []*ChildStruct
	}
	ChildStruct
	GoodKid ChildStruct
	BadKid  *ChildStruct
}

type ChildStruct struct {
	Foobar *ParentStruct
	Inline struct {
		Bar    *int
		Bark   *string
		Coolio *bool
		GrandChildStruct
	}
}

type GrandChildStruct struct {
	Foobar *ChildStruct
	Inline struct {
		Bar *int
	}
}

type ParentStructRecursive struct {
	Inline struct {
		Bar    *int
		Baz    *string
		Coolio *bool
		Child  ChildStructRecursive
		//Children []ChildStruct
		//Nieces   []*ChildStruct
	}
	ChildStructRecursive
	GoodKid ChildStructRecursive
	BadKid  *ChildStructRecursive
}

type ChildStructRecursive struct {
	*ParentStructRecursive
	Inline struct {
		Bar    *int
		Bark   *string
		Coolio *bool
		GrandChildStructRecursive
	}
}

type GrandChildStructRecursive struct {
	*ChildStructRecursive
	Inline struct {
		Bar *int
	}
}

type SomeInterface interface {
	DoThings()
}

type StructWithInterfaceField struct {
	IFace SomeInterface
}

type StructWithEmbeddedInterface struct {
	SomeInterface
}

type SomeInterface1 struct{}

func (SomeInterface1) DoThings() {}

var _ = jsonschema.SetImplementations[SomeInterface](SomeInterface1{})
