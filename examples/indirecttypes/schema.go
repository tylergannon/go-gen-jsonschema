//go:build jsonschema
// +build jsonschema

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

func (MapOfStringToPerson) Schema() json.RawMessage {
	panic("not implemented")
}

func (MapOfStringToPointerToPerson) Schema() json.RawMessage {
	panic("not implemented")
}

func (ComplexStruct) Schema() json.RawMessage {
	panic("not implemented")
}

// Register all the types with the schema generator.
// Each type that needs a schema must be registered here.
var (
	_ = jsonschema.NewJSONSchemaMethod(SimpleInt.Schema)
	_ = jsonschema.NewJSONSchemaMethod(PointerToInt.Schema)
	_ = jsonschema.NewJSONSchemaMethod(PointerToSimpleInt.Schema)
	_ = jsonschema.NewJSONSchemaMethod(SliceOfInt.Schema)
	_ = jsonschema.NewJSONSchemaMethod(SliceOfSimpleInt.Schema)
	_ = jsonschema.NewJSONSchemaMethod(SliceOfPointerToInt.Schema)
	_ = jsonschema.NewJSONSchemaMethod(SliceOfPointerToSimpleInt.Schema)
	_ = jsonschema.NewJSONSchemaMethod(NamedSliceType.Schema)
	_ = jsonschema.NewJSONSchemaMethod(Person.Schema)
	_ = jsonschema.NewJSONSchemaMethod(PointerToPerson.Schema)
	_ = jsonschema.NewJSONSchemaMethod(SliceOfPerson.Schema)
	_ = jsonschema.NewJSONSchemaMethod(SliceOfPointerToPerson.Schema)
	_ = jsonschema.NewJSONSchemaMethod(MapOfStringToPerson.Schema)
	_ = jsonschema.NewJSONSchemaMethod(MapOfStringToPointerToPerson.Schema)
	_ = jsonschema.NewJSONSchemaMethod(ComplexStruct.Schema)
)
