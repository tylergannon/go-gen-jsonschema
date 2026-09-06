package optionality

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

//go:generate go run ../../polytype gen --pretty --validate

type Detail struct {
	Message string `json:"message"`
}

type Pet interface{ pet() }

type Dog struct {
	Name string `json:"name"`
}

func (Dog) pet() {}

func (d Dog) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind string `json:"type"`
		Name string `json:"name"`
	}{Kind: "Dog", Name: d.Name})
}

type Cat struct {
	Lives int `json:"lives"`
}

func (Cat) pet() {}

func (c Cat) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind  string `json:"type"`
		Lives int    `json:"lives"`
	}{Kind: "Cat", Lives: c.Lives})
}

type Config struct {
	Name       string                      `json:"name"`
	MaxRetries polytype.Optional[int]      `json:"max_retries,omitzero"`
	Nickname   polytype.Optional[string]   `json:"nickname,omitzero"`
	Metadata   polytype.Optional[Detail]   `json:"metadata,omitzero"`
	Backup     polytype.Optional[*Detail]  `json:"backup,omitzero"`
	Tags       polytype.Optional[[]string] `json:"tags,omitzero"`
	Pet        polytype.Optional[Pet]      `json:"pet,omitzero"`
	Timeout    polytype.Nullable[int]      `json:"timeout"`
	Detail     polytype.Nullable[Detail]   `json:"detail"`
}

type Count int

// NumericConfig keeps scalar-width coverage on the public generator path.
type NumericConfig struct {
	Count           Count                      `json:"count"`
	OptionalCount   polytype.Optional[Count]   `json:"optional_count,omitzero"`
	NullableCount   polytype.Nullable[Count]   `json:"nullable_count"`
	Int             int                        `json:"int"`
	OptionalInt     polytype.Optional[int]     `json:"optional_int,omitzero"`
	NullableInt     polytype.Nullable[int]     `json:"nullable_int"`
	Int8            int8                       `json:"int8"`
	OptionalInt8    polytype.Optional[int8]    `json:"optional_int8,omitzero"`
	NullableInt8    polytype.Nullable[int8]    `json:"nullable_int8"`
	Int16           int16                      `json:"int16"`
	OptionalInt16   polytype.Optional[int16]   `json:"optional_int16,omitzero"`
	NullableInt16   polytype.Nullable[int16]   `json:"nullable_int16"`
	Int32           int32                      `json:"int32"`
	OptionalInt32   polytype.Optional[int32]   `json:"optional_int32,omitzero"`
	NullableInt32   polytype.Nullable[int32]   `json:"nullable_int32"`
	Int64           int64                      `json:"int64"`
	OptionalInt64   polytype.Optional[int64]   `json:"optional_int64,omitzero"`
	NullableInt64   polytype.Nullable[int64]   `json:"nullable_int64"`
	Uint            uint                       `json:"uint"`
	OptionalUint    polytype.Optional[uint]    `json:"optional_uint,omitzero"`
	NullableUint    polytype.Nullable[uint]    `json:"nullable_uint"`
	Uint8           uint8                      `json:"uint8"`
	OptionalUint8   polytype.Optional[uint8]   `json:"optional_uint8,omitzero"`
	NullableUint8   polytype.Nullable[uint8]   `json:"nullable_uint8"`
	Uint16          uint16                     `json:"uint16"`
	OptionalUint16  polytype.Optional[uint16]  `json:"optional_uint16,omitzero"`
	NullableUint16  polytype.Nullable[uint16]  `json:"nullable_uint16"`
	Uint32          uint32                     `json:"uint32"`
	OptionalUint32  polytype.Optional[uint32]  `json:"optional_uint32,omitzero"`
	NullableUint32  polytype.Nullable[uint32]  `json:"nullable_uint32"`
	Uint64          uint64                     `json:"uint64"`
	OptionalUint64  polytype.Optional[uint64]  `json:"optional_uint64,omitzero"`
	NullableUint64  polytype.Nullable[uint64]  `json:"nullable_uint64"`
	Float32         float32                    `json:"float32"`
	OptionalFloat32 polytype.Optional[float32] `json:"optional_float32,omitzero"`
	NullableFloat32 polytype.Nullable[float32] `json:"nullable_float32"`
	Float64         float64                    `json:"float64"`
	OptionalFloat64 polytype.Optional[float64] `json:"optional_float64,omitzero"`
	NullableFloat64 polytype.Nullable[float64] `json:"nullable_float64"`
}
