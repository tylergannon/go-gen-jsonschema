//go:build jsonschema

package indirecttypes

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

// Schema methods for all the types we want to generate schemas for.
// Each method is a stub that will be replaced during code generation.

func (SimpleInt) Schema() json.RawMessage {
	panic("not implemented")
}

func (PointerToInt) Schema() json.RawMessage {
	panic("not implemented")
}

func (PointerToSimpleInt) Schema() json.RawMessage {
	panic("not implemented")
}

func (SliceOfInt) Schema() json.RawMessage {
	panic("not implemented")
}

func (SliceOfSimpleInt) Schema() json.RawMessage {
	panic("not implemented")
}

func (SliceOfPointerToInt) Schema() json.RawMessage {
	panic("not implemented")
}

func (SliceOfPointerToSimpleInt) Schema() json.RawMessage {
	panic("not implemented")
}

func (NamedSliceType) Schema() json.RawMessage {
	panic("not implemented")
}

func (Person) Schema() json.RawMessage {
	panic("not implemented")
}

func (PointerToPerson) Schema() json.RawMessage {
	panic("not implemented")
}

func (SliceOfPerson) Schema() json.RawMessage {
	panic("not implemented")
}

func (SliceOfPointerToPerson) Schema() json.RawMessage {
	panic("not implemented")
}

// COMMENTED OUT: Map types are not yet supported
// func (MapOfStringToPerson) Schema() json.RawMessage {
// 	panic("not implemented")
// }

// func (MapOfStringToPointerToPerson) Schema() json.RawMessage {
// 	panic("not implemented")
// }

func (ComplexStruct) Schema() json.RawMessage {
	panic("not implemented")
}

// Register all the types with the schema generator.
// Each type that needs a schema must be registered here.
var (
	_ = jsonschema.Declare(SimpleInt.Schema)
	_ = jsonschema.Declare(PointerToInt.Schema)
	_ = jsonschema.Declare(PointerToSimpleInt.Schema)
	_ = jsonschema.Declare(SliceOfInt.Schema)
	_ = jsonschema.Declare(SliceOfSimpleInt.Schema)
	_ = jsonschema.Declare(SliceOfPointerToInt.Schema)
	_ = jsonschema.Declare(SliceOfPointerToSimpleInt.Schema)
	_ = jsonschema.Declare(NamedSliceType.Schema)
	_ = jsonschema.Declare(Person.Schema)
	_ = jsonschema.Declare(PointerToPerson.Schema)
	_ = jsonschema.Declare(SliceOfPerson.Schema)
	_ = jsonschema.Declare(SliceOfPointerToPerson.Schema)
	// COMMENTED OUT: Map types are not yet supported
	// _ = jsonschema.Declare(MapOfStringToPerson.Schema)
	// _ = jsonschema.Declare(MapOfStringToPointerToPerson.Schema)
	_ = jsonschema.Declare(ComplexStruct.Schema)
)
