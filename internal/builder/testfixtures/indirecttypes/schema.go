//go:build jsonschema

package basictypes

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (IntType) Schema() json.RawMessage {
	panic("not implemented")
}

func (PointerToIntType) Schema() json.RawMessage {
	panic("not implemented")
}

func (PointerToNamedType) Schema() json.RawMessage {
	panic("not implemented")
}

func (DefinedAsNamedType) Schema() json.RawMessage {
	panic("not implemented")
}

func (SliceOfPointerToInt) Schema() json.RawMessage {
	panic("not implemented")
}

func (SliceOfPointerToNamedType) Schema() json.RawMessage {
	panic("not implemented")
}

func (SliceOfNamedType) Schema() json.RawMessage {
	panic("not implemented")
}

func (NamedSliceType) Schema() json.RawMessage {
	panic("not implemented")
}

func (NamedNamedSliceType) Schema() json.RawMessage {
	panic("not implemented")
}

func (SliceOfNamedNamedSliceType) Schema() json.RawMessage {
	panic("not implemented")
}

func (PointerToRemoteType) Schema() json.RawMessage {
	panic("not implemented")
}

func (DefinedAsRemoteType) Schema() json.RawMessage {
	panic("not implemented")
}

func (DefinedAsRemoteSliceType) Schema() json.RawMessage {
	panic("not implemented")
}

func (DefinedAsPointerToRemoteSliceType) Schema() json.RawMessage {
	panic("not implemented")
}

func (DefinedAsSliceOfRemoteSliceType) Schema() json.RawMessage {
	panic("not implemented")
}

var (
	_ = polytype.NewJSONSchemaMethod(IntType.Schema)
	_ = polytype.NewJSONSchemaMethod(PointerToIntType.Schema)
	_ = polytype.NewJSONSchemaMethod(PointerToNamedType.Schema)
	_ = polytype.NewJSONSchemaMethod(DefinedAsNamedType.Schema)
	_ = polytype.NewJSONSchemaMethod(SliceOfPointerToInt.Schema)
	_ = polytype.NewJSONSchemaMethod(SliceOfPointerToNamedType.Schema)
	_ = polytype.NewJSONSchemaMethod(SliceOfNamedType.Schema)
	_ = polytype.NewJSONSchemaMethod(NamedSliceType.Schema)
	_ = polytype.NewJSONSchemaMethod(NamedNamedSliceType.Schema)
	_ = polytype.NewJSONSchemaMethod(SliceOfNamedNamedSliceType.Schema)
	_ = polytype.NewJSONSchemaMethod(PointerToRemoteType.Schema)
	_ = polytype.NewJSONSchemaMethod(DefinedAsRemoteType.Schema)
	_ = polytype.NewJSONSchemaMethod(DefinedAsRemoteSliceType.Schema)
	_ = polytype.NewJSONSchemaMethod(DefinedAsPointerToRemoteSliceType.Schema)
	_ = polytype.NewJSONSchemaMethod(DefinedAsSliceOfRemoteSliceType.Schema)
)
