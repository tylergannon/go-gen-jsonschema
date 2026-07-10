package optionality

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

//go:generate go run ../../gen-jsonschema gen --pretty --validate

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
		Kind string `json:"!kind"`
		Name string `json:"name"`
	}{Kind: "Dog", Name: d.Name})
}

type Cat struct {
	Lives int `json:"lives"`
}

func (Cat) pet() {}

func (c Cat) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind  string `json:"!kind"`
		Lives int    `json:"lives"`
	}{Kind: "Cat", Lives: c.Lives})
}

type Config struct {
	Name       string                        `json:"name"`
	MaxRetries jsonschema.Optional[int]      `json:"max_retries,omitzero"`
	Nickname   jsonschema.Optional[string]   `json:"nickname,omitzero"`
	Metadata   jsonschema.Optional[Detail]   `json:"metadata,omitzero"`
	Backup     jsonschema.Optional[*Detail]  `json:"backup,omitzero"`
	Tags       jsonschema.Optional[[]string] `json:"tags,omitzero"`
	Pet        jsonschema.Optional[Pet]      `json:"pet,omitzero"`
	Timeout    jsonschema.Nullable[int]      `json:"timeout"`
	Detail     jsonschema.Nullable[Detail]   `json:"detail"`
}

type Count int

// NumericConfig keeps scalar-width coverage on the public generator path.
type NumericConfig struct {
	Count           Count                        `json:"count"`
	OptionalCount   jsonschema.Optional[Count]   `json:"optional_count,omitzero"`
	NullableCount   jsonschema.Nullable[Count]   `json:"nullable_count"`
	Int             int                          `json:"int"`
	OptionalInt     jsonschema.Optional[int]     `json:"optional_int,omitzero"`
	NullableInt     jsonschema.Nullable[int]     `json:"nullable_int"`
	Int8            int8                         `json:"int8"`
	OptionalInt8    jsonschema.Optional[int8]    `json:"optional_int8,omitzero"`
	NullableInt8    jsonschema.Nullable[int8]    `json:"nullable_int8"`
	Int16           int16                        `json:"int16"`
	OptionalInt16   jsonschema.Optional[int16]   `json:"optional_int16,omitzero"`
	NullableInt16   jsonschema.Nullable[int16]   `json:"nullable_int16"`
	Int32           int32                        `json:"int32"`
	OptionalInt32   jsonschema.Optional[int32]   `json:"optional_int32,omitzero"`
	NullableInt32   jsonschema.Nullable[int32]   `json:"nullable_int32"`
	Int64           int64                        `json:"int64"`
	OptionalInt64   jsonschema.Optional[int64]   `json:"optional_int64,omitzero"`
	NullableInt64   jsonschema.Nullable[int64]   `json:"nullable_int64"`
	Uint            uint                         `json:"uint"`
	OptionalUint    jsonschema.Optional[uint]    `json:"optional_uint,omitzero"`
	NullableUint    jsonschema.Nullable[uint]    `json:"nullable_uint"`
	Uint8           uint8                        `json:"uint8"`
	OptionalUint8   jsonschema.Optional[uint8]   `json:"optional_uint8,omitzero"`
	NullableUint8   jsonschema.Nullable[uint8]   `json:"nullable_uint8"`
	Uint16          uint16                       `json:"uint16"`
	OptionalUint16  jsonschema.Optional[uint16]  `json:"optional_uint16,omitzero"`
	NullableUint16  jsonschema.Nullable[uint16]  `json:"nullable_uint16"`
	Uint32          uint32                       `json:"uint32"`
	OptionalUint32  jsonschema.Optional[uint32]  `json:"optional_uint32,omitzero"`
	NullableUint32  jsonschema.Nullable[uint32]  `json:"nullable_uint32"`
	Uint64          uint64                       `json:"uint64"`
	OptionalUint64  jsonschema.Optional[uint64]  `json:"optional_uint64,omitzero"`
	NullableUint64  jsonschema.Nullable[uint64]  `json:"nullable_uint64"`
	Float32         float32                      `json:"float32"`
	OptionalFloat32 jsonschema.Optional[float32] `json:"optional_float32,omitzero"`
	NullableFloat32 jsonschema.Nullable[float32] `json:"nullable_float32"`
	Float64         float64                      `json:"float64"`
	OptionalFloat64 jsonschema.Optional[float64] `json:"optional_float64,omitzero"`
	NullableFloat64 jsonschema.Nullable[float64] `json:"nullable_float64"`
}
