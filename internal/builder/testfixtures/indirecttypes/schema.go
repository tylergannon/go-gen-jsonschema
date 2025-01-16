//go:build jsonschema
// +build jsonschema

package basictypes

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (IntType) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (PointerToIntType) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (PointerToNamedType) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (DefinedAsNamedType) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (SliceOfPointerToInt) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (SliceOfPointerToNamedType) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (SliceOfNamedType) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (NamedSliceType) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (NamedNamedSliceType) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (SliceOfNamedNamedSliceType) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (PointerToRemoteType) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (DefinedAsRemoteType) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (DefinedAsRemoteSliceType) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (DefinedAsPointerToRemoteSliceType) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (DefinedAsSliceOfRemoteSliceType) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

var (
	_ = jsonschema.NewJSONSchemaMethod(IntType.Schema)
	_ = jsonschema.NewJSONSchemaMethod(PointerToIntType.Schema)
	_ = jsonschema.NewJSONSchemaMethod(PointerToNamedType.Schema)
	_ = jsonschema.NewJSONSchemaMethod(DefinedAsNamedType.Schema)
	_ = jsonschema.NewJSONSchemaMethod(SliceOfPointerToInt.Schema)
	_ = jsonschema.NewJSONSchemaMethod(SliceOfPointerToNamedType.Schema)
	_ = jsonschema.NewJSONSchemaMethod(SliceOfNamedType.Schema)
	_ = jsonschema.NewJSONSchemaMethod(NamedSliceType.Schema)
	_ = jsonschema.NewJSONSchemaMethod(NamedNamedSliceType.Schema)
	_ = jsonschema.NewJSONSchemaMethod(SliceOfNamedNamedSliceType.Schema)
	_ = jsonschema.NewJSONSchemaMethod(PointerToRemoteType.Schema)
	_ = jsonschema.NewJSONSchemaMethod(DefinedAsRemoteType.Schema)
	_ = jsonschema.NewJSONSchemaMethod(DefinedAsRemoteSliceType.Schema)
	_ = jsonschema.NewJSONSchemaMethod(DefinedAsPointerToRemoteSliceType.Schema)
	_ = jsonschema.NewJSONSchemaMethod(DefinedAsSliceOfRemoteSliceType.Schema)
)
