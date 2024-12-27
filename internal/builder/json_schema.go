package builder

import (
	"encoding/json"
	"fmt"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry"
	"go/types"
	"reflect"
	"sort"
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
		"anyOf": mapSlice(j, func(it *jsonSchema, _ int) json.Marshaler { return json.Marshaler(it) }),
	}
	return json.Marshal(toMarshal)
}

var _ json.Marshaler = jsonUnionType{}

type RefElement []byte

func (r RefElement) MarshalJSON() ([]byte, error) {
	return r, nil
}

// A ref into definitions
func refElement(ref string) RefElement {
	return RefElement(fmt.Sprintf(`{"$ref": "%s"}`, ref))
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

var _ json.Marshaler = (*jsonSchema)(nil)

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
				return nil, fmt.Errorf("failed to marshal property %s of type %T: %w", p.name, p.def, err)
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
		b.WriteRune('"')
		b.WriteString(definitionsKey)
		b.WriteRune('"')
		b.WriteRune(':')
		defs := basicMarshaler(j.Definitions)
		defsData, err := json.Marshal(defs)
		if err != nil {
			return nil, err
		}
		b.Write(defsData)
		b.WriteRune(',')
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

func constElement[T ~int | ~string | ~bool](val T) basicMarshaler {
	_val, _ := json.Marshal(val)
	it := basicMarshaler{"const": json.RawMessage(_val)}
	var typ = reflect.TypeFor[T]()
	if typ.Kind() == reflect.Bool {
		it["type"] = json.RawMessage(`"boolean"`)
	} else if typ.Kind() == reflect.String {
		it["type"] = json.RawMessage(`"string"`)
	} else if typ.Kind() == reflect.Int {
		it["type"] = json.RawMessage(`"integer"`)
	} else {
		panic(fmt.Sprintf("unsupported type for const element %T", val))
	}
	return it
}

func newEnumType(node typeregistry.EnumTypeNode) json.Marshaler {
	var (
		sb               = strings.Builder{}
		values           = make([]string, len(node.Entries))
		haveDescriptions bool
		comments         = make([]string, len(node.Entries))
	)
	sb.WriteString(buildComments(node.NamedTypeNode.TypeSpec.Decorations()))

	for i, val := range node.Entries {
		if val.Decorations != nil {
			comments[i] = strings.TrimSpace(strings.TrimPrefix(buildComments(val.Decorations), val.Name))
		}
		values[i] = val.Value
		if len(comments[i]) > 0 {
			haveDescriptions = true
		}
	}
	if haveDescriptions && sb.Len() > 0 {
		sb.WriteString("\n\n")
	}
	if haveDescriptions {
		sb.WriteString("## Values\n\n")
	}
	written := 0
	for i, val := range values {
		if len(comments[i]) == 0 {
			continue
		}
		if written > 0 {
			sb.WriteString("\n\n")
		}
		written++
		sb.WriteString("### ")
		sb.WriteString(val)
		sb.WriteString("\n\n")
		sb.WriteString(comments[i])
	}
	var (
		result = basicMarshaler{
			"type": json.RawMessage(`"string"`),
		}
		valuesBytes, _ = json.Marshal(values)
	)
	result["enum"] = json.RawMessage(valuesBytes)
	if sb.Len() > 0 {
		var (
			descriptionBytes, _ = json.Marshal(sb.String())
		)
		result["description"] = json.RawMessage(descriptionBytes)
	}
	return result
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
		"type": rawString(jsonSchemaDataTypeName),
	}
}

func arraySchema(items json.Marshaler, description string) basicMarshaler {
	var res = basicMarshaler{
		"type":  rawString("array"),
		"items": items,
	}
	if description != "" {
		res["description"] = rawString(description)
	}
	return res
}

func stringSchema(description string) basicMarshaler {
	return basicMarshaler{
		"type":        rawString(String),
		"description": rawString(description),
	}
}

func boolSchema(description string) basicMarshaler {
	return basicMarshaler{
		"type":        rawString(Boolean),
		"description": rawString(description),
	}
}

func rawString[T ~string](s T) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`"%s"`, s))
}
func intSchema(description string) basicMarshaler {
	return basicMarshaler{
		"type":        rawString(Integer),
		"description": rawString(description),
	}
}

type basicMarshaler map[string]json.Marshaler

// MarshalJSON implements json.Marshaler.
func (b basicMarshaler) MarshalJSON() ([]byte, error) {
	// Collect and sort keys
	keys := make([]string, 0, len(b))
	for k := range b {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Manually assemble the JSON object
	var sb strings.Builder
	sb.WriteByte('{')

	for i, k := range keys {
		// Add a comma if this is not the first item
		if i != 0 {
			sb.WriteByte(',')
		}
		// Marshal the key as a JSON string
		keyData, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		sb.Write(keyData)
		sb.WriteByte(':')

		// Marshal the value using each element's MarshalJSON
		valData, err := b[k].MarshalJSON()
		if err != nil {
			return nil, err
		}
		sb.Write(valData)

	}

	sb.WriteByte('}')
	return []byte(sb.String()), nil
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
