package builder

import (
	"encoding/json"
	"fmt"
	"go/types"
	"strconv"
	"strings"
)

type (
	dataType string
)

const (
	defaultDefinitionsKey = "$defs"
)

const (
	Object  dataType = "object"
	Number  dataType = "number"
	Integer dataType = "integer"
	String  dataType = "string"
	Array   dataType = "array"
	Null    dataType = "null"
	Boolean dataType = "boolean"
)

type jsonUnionType []*jsonSchema

// MarshalJSON implements json.Marshaler.
func (j jsonUnionType) MarshalJSON() ([]byte, error) {
	toMarshal := map[string][]json.Marshaler{
		"anyOf": mapSlice([]*jsonSchema(j), func(it *jsonSchema, _ int) json.Marshaler { return json.Marshaler(it) }),
	}
	return json.Marshal(toMarshal)
}

var _ json.Marshaler = jsonUnionType{}

// An anyOf element
func unionSchemaElement(alts ...json.Marshaler) json.Marshaler {
	return basicMarshaler{
		"anyOf": alts,
	}
}

// A ref into definitions
func refElement(ref string) json.Marshaler {
	return basicMarshaler{"$ref": ref}
}

type schemaProperty struct {
	name string
	def  json.Marshaler
}

// jsonSchema is a struct for describing a JSON Schema. It is fairly limited,
// and you may have better luck using a third-party library. This is a copy from
// go-openai's "jsonschema.Definition{}" struct, with the difference being that
// this one holds references to json.Marshaler, rather than to itself.
type jsonSchema struct {
	// Description is the description of the schema.
	Description string
	// Properties describes the properties of an object, if the schema type is
	// Object.
	Properties []schemaProperty
	// Required specifies which properties are required, if the schema type is
	// Object.
	Required []string
	// Items specifies which data type an array contains, if the schema type is
	// Array.
	AdditionalProperties any
	Definitions          map[string]json.Marshaler
	Strict               bool
	// DefinitionsKey is the name of the key to use for writing definitions,
	// if present.
	// Defaults to "definitions"
	DefinitionsKey string
}

func (j *jsonSchema) MarshalJSON() ([]byte, error) {
	var b strings.Builder
	b.WriteByte('{')

	// "description"
	if j.Description != "" {
		b.WriteString(`"description":`)
		enc, err := json.Marshal(j.Description)
		if err != nil {
			return nil, err
		}
		b.Write(enc)
		b.WriteByte(',')
	}

	// "type"
	b.WriteString(`"type":"object",`)

	// "properties"
	if len(j.Properties) > 0 {
		b.WriteString(`"properties":{`)
		for i, p := range j.Properties {
			encProp, err := p.def.MarshalJSON()
			if err != nil {
				return nil, err
			}
			b.WriteString(strconv.Quote(p.name))
			b.WriteByte(':')
			b.Write(encProp)
			if i < len(j.Properties)-1 {
				b.WriteByte(',')
			}
		}
		b.WriteByte('}')
		b.WriteByte(',')
	}

	// If Strict, "required" = all property names, "additionalProperties"=false
	if j.Strict {
		b.WriteString(`"required":[`)
		for i, p := range j.Properties {
			b.WriteString(strconv.Quote(p.name))
			if i < len(j.Properties)-1 {
				b.WriteByte(',')
			}
		}
		b.WriteString(`],"additionalProperties":false,`)
	} else if len(j.Required) > 0 {
		// If not strict but required is explicitly given
		b.WriteString(`"required":[`)
		for i, r := range j.Required {
			b.WriteString(strconv.Quote(r))
			if i < len(j.Required)-1 {
				b.WriteByte(',')
			}
		}
		b.WriteString(`],`)
	}

	// "definitions"
	if len(j.Definitions) > 0 {
		var definitionsKey = defaultDefinitionsKey
		if j.DefinitionsKey != "" {
			definitionsKey = j.DefinitionsKey
		}
		b.WriteByte('"')
		b.WriteString(definitionsKey)
		b.WriteString(`":{`)
		i := 0
		for k, v := range j.Definitions {
			encDef, err := v.MarshalJSON()
			if err != nil {
				return nil, err
			}
			b.WriteString(strconv.Quote(k))
			b.WriteByte(':')
			b.Write(encDef)
			if i < len(j.Definitions)-1 {
				b.WriteByte(',')
			}
			i++
		}
		b.WriteString(`},`)
	}

	// If not strict, "additionalProperties" if set
	if !j.Strict && j.AdditionalProperties != nil {
		b.WriteString(`"additionalProperties":`)
		encAP, err := json.Marshal(j.AdditionalProperties)
		if err != nil {
			return nil, err
		}
		b.Write(encAP)
		b.WriteByte(',')
	}

	// Trim trailing comma if any
	out := b.String()
	if out[len(out)-1] == ',' {
		out = out[:len(out)-1]
	}
	out += "}"
	return []byte(out), nil
}

func constSchema[T ~int | ~string | ~bool](val T, description string) basicMarshaler {
	var schemaType dataType
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

func enumSchema[T ~int | ~string](description string, vals ...T) basicMarshaler {
	var schemaType dataType
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

func newBasicType(t *types.Basic) json.Marshaler {
	var jsonSchemaDataTypeName string
	switch t.Kind() {
	case types.String:
		jsonSchemaDataTypeName = "string"
	case types.Bool:
		jsonSchemaDataTypeName = "boolean"
	case types.Int:
		jsonSchemaDataTypeName = "integer"
	case types.Float32, types.Float64:
		jsonSchemaDataTypeName = "number"
	case types.Int8, types.Int16, types.Int32, types.Int64, types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr:
		jsonSchemaDataTypeName = "integer"
	default:
		panic(fmt.Sprintf("unknown type kind: %v", t.Kind()))
	}
	return basicMarshaler{
		"type": jsonSchemaDataTypeName,
	}
}

func arraySchema(items json.Marshaler, description string) basicMarshaler {
	var res = basicMarshaler{
		"type":  "array",
		"items": items,
	}
	if description != "" {
		res["description"] = description
	}
	return res
}

func stringSchema(description string) basicMarshaler {
	return basicMarshaler{
		"type":        String,
		"description": description,
	}
}

func boolSchema(description string) basicMarshaler {
	return basicMarshaler{
		"type":        Boolean,
		"description": description,
	}
}

func intSchema(description string) basicMarshaler {
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

// mapSlice manipulates a slice and transforms it to a slice of another type.
// Play: https://go.dev/play/p/OkPcYAhBo0D
func mapSlice[T any, R any](collection []T, iteratee func(item T, index int) R) []R {
	result := make([]R, len(collection))

	for i := range collection {
		result[i] = iteratee(collection[i], i)
	}

	return result
}