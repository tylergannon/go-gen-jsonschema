package jsonschema

import (
	"encoding/json"
	"errors"
)

type DataType string

const (
	Object  DataType = "object"
	Number  DataType = "number"
	Integer DataType = "integer"
	String  DataType = "string"
	Array   DataType = "array"
	Null    DataType = "null"
	Boolean DataType = "boolean"
)

type JSONUnionType []*JSONSchema

// MarshalJSON implements json.Marshaler.
func (j JSONUnionType) MarshalJSON() ([]byte, error) {
	var marshalers []json.Marshaler
	for _, it := range j {
		marshalers = append(marshalers, it)
	}
	toMarshal := map[string][]json.Marshaler{
		"anyOf": marshalers,
	}
	return json.Marshal(toMarshal)
}

var _ json.Marshaler = JSONUnionType{}

// An anyOf element
func UnionSchemaEl(alts ...json.Marshaler) json.Marshaler {
	return basicMarshaler{
		"anyOf": alts,
	}
}

// A ref into definitions
func RefSchemaEl(ref string) json.Marshaler {
	return basicMarshaler{"$ref": ref}
}

// JSONSchema is a struct for describing a JSON Schema. It is fairly limited,
// and you may have better luck using a third-party library. This is a copy from
// go-openai's "jsonschema.Definition{}" struct, with the difference being that
// this one holds references to json.Marshaler, rather than to itself.
type JSONSchema struct {
	// Type specifies the data type of the schema.
	Type DataType `json:"type" yaml:"type"`
	// Description is the description of the schema.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Enum is used to restrict a value to a fixed set of values. It must be an
	// array with at least one element, where each element is unique. You will
	// probably only use this with strings.
	Enum []any `json:"enum,omitempty" yaml:"enum,omitempty"`
	// Properties describes the properties of an object, if the schema type is
	// Object.
	Properties map[string]json.Marshaler `json:"properties,omitempty" yaml:"properties,omitempty"`
	// Required specifies which properties are required, if the schema type is
	// Object.
	Required []string `json:"required,omitempty" yaml:"required,omitempty"`
	// Items specifies which data type an array contains, if the schema type is
	// Array.
	Items json.Marshaler `json:"items,omitempty" yaml:"items,omitempty"`
	// AdditionalProperties is used to control the handling of properties in an
	// object that are not explicitly defined in the properties section of the
	// schema. example: additionalProperties: true additionalProperties: false
	// additionalProperties: jsonschema.Definition{Type: jsonschema.String}
	AdditionalProperties any                       `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Definitions          map[string]json.Marshaler `json:"$defs,omitzero" yaml:"$defs,omitempty"`
	Const                any                       `json:"const,omitempty"` // Provide a const value
	// Strict will make all properties required and additionalProperties: false if
	// not already set. pplies only if Type = "object".
	Strict bool `json:"-" yaml:"-"`
}

func (s JSONSchema) MarshalJSON() ([]byte, error) {
	type Alias JSONSchema
	var v = Alias(s)
	if v.Type == "" {
		return nil, errors.New("type is required")
	}
	if v.Type == Object && v.Properties == nil {
		v.Properties = make(map[string]json.Marshaler)
	}
	if s.Strict && v.Type == Object {
		if v.AdditionalProperties == nil {
			v.AdditionalProperties = false
		}
		v.Required = make([]string, 0, len(s.Properties))
		for k := range s.Properties {
			v.Required = append(v.Required, k)
		}
	}
	return json.Marshal(v)
}

var _ json.Marshaler = JSONSchema{}

func ConstSchema[T ~int | ~string](val T, description string) json.Marshaler {
	var schemaType DataType
	if _, ok := any(val).(int); ok {
		schemaType = "integer"
	} else {
		schemaType = "string"
	}

	res := basicMarshaler{
		"type":  schemaType,
		"const": val,
	}
	if description != "" {
		res["description"] = description
	}
	return res
}

func EnumSchema[T ~int | ~string](description string, vals ...T) json.Marshaler {
	var schemaType DataType
	if _, ok := any(vals[0]).(int); ok {
		schemaType = "integer"
	} else {
		schemaType = "string"
	}
	res := basicMarshaler{
		"type": schemaType,
		"enum": vals,
	}
	if description != "" {
		res["description"] = description
	}
	return res
}

func ArraySchema(items json.Marshaler, description string) json.Marshaler {
	var res = basicMarshaler{
		"type":  "array",
		"items": items,
	}
	if description != "" {
		res["description"] = description
	}
	return res
}

func StringSchema(description string) json.Marshaler {
	return basicMarshaler{
		"type":        String,
		"description": description,
	}
}

func BoolSchema(description string) json.Marshaler {
	return basicMarshaler{
		"type":        Boolean,
		"description": description,
	}
}

func IntSchema(description string) json.Marshaler {
	return basicMarshaler{
		"type":        Integer,
		"description": description,
	}
}

type basicMarshaler map[string]any

// MarshalJSON implements json.Marshaler.
func (b basicMarshaler) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any(b))
}

var _ json.Marshaler = basicMarshaler{}
