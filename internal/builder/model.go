package builder

import (
	"encoding/json"
	"fmt"
	"github.com/tylergannon/go-gen-jsonschema/internal/scanner"
	"strconv"
	"strings"
)

const (
	DefaultDiscriminatorPropName = "!type"
)

type (
	JSONSchema interface {
		json.Marshaler
		jsonSchemaMarker()
		TypeID() scanner.TypeID
	}

	schemaNode interface {
		Type() string
		Description() string
		SetDescription(desc string)
		JSONSchema
	}

	// ObjectProp represents a single property in an ObjectNode.
	ObjectProp struct {
		Name     string
		Schema   JSONSchema
		Optional bool
	}

	// ObjectNode represents an object schema.
	// Discriminator: always non-empty, but only used when included in a union (anyOf).
	ObjectNode struct {
		Desc          string         `json:"description,omitempty"`
		Properties    []ObjectProp   `json:"properties,omitempty"`
		Discriminator string         `json:"-"`
		TypeID_       scanner.TypeID `json:"-"`
	}

	// PropertyNode is a scalar property (string, int, bool).
	//   - `Const` is an exact value that the field must match.
	//   - `Enum` is an array of allowable values.
	//   - If both `Const` and `Enum` are set, the field effectively has a single valid value (the `Const`) plus whatever is in `Enum`—though that’s unusual in practice.
	PropertyNode[T ~int | ~string | ~bool | float32 | float64] struct {
		Desc    string         `json:"description,omitempty"`
		Enum    []T            `json:"enum,omitempty"`
		Const   *T             `json:"const,omitempty"`
		Typ     string         `json:"type,omitempty"`
		TypeID_ scanner.TypeID `json:"-"`
	}

	ConstNode[T ~int | ~string | ~bool | float32 | float64] struct {
		PropertyNode[T]
		Const T `json:"const"`
	}

	ArrayNode struct {
		Desc    string         `json:"description,omitempty"`
		Items   JSONSchema     `json:"items,omitempty"`
		TypeID_ scanner.TypeID `json:"-"`
	}

	// UnionTypeNode means `{"anyOf": [ <object1-with-discriminator>, ... ]}`.
	UnionTypeNode struct {
		DiscriminatorPropName string
		Options               []ObjectNode
		TypeID_               scanner.TypeID `json:"-"`
	}
)

//---------------------------------------------------------------------
// Ensure each node satisfies the schemaNode or JSONSchema interface
//---------------------------------------------------------------------

var (
	_ JSONSchema = UnionTypeNode{}
	_ schemaNode = ArrayNode{}
	_ schemaNode = PropertyNode[int]{}
	_ schemaNode = ObjectNode{}
)

//---------------------------------------------------------------------
// ObjectNode
//---------------------------------------------------------------------

func (o ObjectNode) TypeID() scanner.TypeID { return o.TypeID_ }

func (o ObjectNode) Type() string {
	return "object"
}

func (o ObjectNode) Description() string {
	return o.Desc
}

func (o ObjectNode) jsonSchemaMarker() {}

func (o ObjectNode) SetDescription(s string) {
	o.Desc = s
}

// MarshalJSON for an ObjectNode does NOT embed the Discriminator property
// unless it's included in a UnionTypeNode.
func (o ObjectNode) MarshalJSON() ([]byte, error) {
	var sb strings.Builder
	sb.WriteByte('{')

	// 1. "type":"object"
	sb.WriteString(`"type":"object"`)

	// 2. "description"
	if o.Desc != "" {
		sb.WriteString(`,"description":`)
		encodeString(&sb, o.Desc)
	}

	// 3. "properties"
	if len(o.Properties) > 0 {
		sb.WriteString(`,"properties":{`)
		for i, prop := range o.Properties {
			if i > 0 {
				sb.WriteByte(',')
			}
			encodeString(&sb, prop.Name)
			sb.WriteByte(':')

			data, err := prop.Schema.MarshalJSON()
			if err != nil {
				return nil, fmt.Errorf("object property %q: %w", prop.Name, err)
			}
			sb.Write(data)
		}
		sb.WriteByte('}')
	}

	// 4. "required"
	requiredFields := make([]string, 0, len(o.Properties))
	for _, prop := range o.Properties {
		if !prop.Optional {
			requiredFields = append(requiredFields, prop.Name)
		}
	}
	if len(requiredFields) > 0 {
		sb.WriteString(`,"required":[`)
		for i, rf := range requiredFields {
			if i > 0 {
				sb.WriteByte(',')
			}
			encodeString(&sb, rf)
		}
		sb.WriteByte(']')
	}

	sb.WriteByte('}')
	return []byte(sb.String()), nil
}

//---------------------------------------------------------------------
// PropertyNode[T]
//---------------------------------------------------------------------

func (p PropertyNode[T]) TypeID() scanner.TypeID { return p.TypeID_ }

func (p PropertyNode[T]) Type() string {
	return p.Typ
}

func (p PropertyNode[T]) Description() string {
	return p.Desc
}

func (p PropertyNode[T]) jsonSchemaMarker() {}

func (p PropertyNode[T]) SetDescription(s string) {
	p.Desc = s
}

// Sample order: type -> description -> const -> enum
func (p PropertyNode[T]) MarshalJSON() ([]byte, error) {
	var sb strings.Builder
	sb.WriteByte('{')

	// 1. "type"
	sb.WriteString(`"type":`)
	encodeString(&sb, p.Typ)

	// 2. "description"
	if p.Desc != "" {
		sb.WriteString(`,"description":`)
		encodeString(&sb, p.Desc)
	}

	// 3. "const"
	// We always output "const" even if it's zero-like.
	// If you want to skip zero-values, you'd need a separate sentinel or pointer.
	// We'll do a quick test if T is zero or not, but that might be insufficient if T=0 is a legit const.
	// So let's always write "const" if p.Const differs from the default generic or if the user intended it:
	constVal, isConst := toJSONValue(p.Const)
	// We'll treat "zero" as valid. If you truly want to skip it, you'd do a custom approach.
	if isConst {
		sb.WriteString(`,"const":`)
		sb.WriteString(constVal)
	}

	// 4. "enum"
	if len(p.Enum) > 0 {
		sb.WriteString(`,"enum":[`)
		for i, val := range p.Enum {
			if i > 0 {
				sb.WriteByte(',')
			}
			strVal, _ := toJSONValue(&val)
			sb.WriteString(strVal)
		}
		sb.WriteByte(']')
	}

	sb.WriteByte('}')
	return []byte(sb.String()), nil
}

// toJSONValue returns a JSON literal for a T (~int|~string|~bool).
// Also returns a bool indicating if we consider this a “valid” value (always true here).
func toJSONValue[T ~int | ~string | ~bool | float64 | float32](v *T) (string, bool) {
	if v == nil {
		return "", false
	}
	var val = *v
	switch u := any(val).(type) {
	case string:
		// JSON-escape the string
		b, _ := json.Marshal(any(u).(string))
		return string(b), true
	case bool:
		return strconv.FormatBool(any(u).(bool)), true
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", u), true
	default:
		panic(fmt.Sprintf("unknown type to make into JSON value %T %#v", u, u))
	}
}

// ---------------------------------------------------------------------
// ArrayNode
// ---------------------------------------------------------------------
func (a ArrayNode) SetDescription(s string) {
	a.Desc = s
}
func (a ArrayNode) TypeID() scanner.TypeID { return a.TypeID_ }

func (a ArrayNode) Type() string {
	return "array"
}

func (a ArrayNode) Description() string {
	return a.Desc
}

func (a ArrayNode) jsonSchemaMarker() {}

// Marshal as:
//
//	{
//	  "type":"array",
//	  "description":"...",
//	  "items": ...
//	}
func (a ArrayNode) MarshalJSON() ([]byte, error) {
	var sb strings.Builder
	sb.WriteByte('{')

	// "type":"array"
	sb.WriteString(`"type":"array"`)

	// "description"
	if a.Desc != "" {
		sb.WriteString(`,"description":`)
		encodeString(&sb, a.Desc)
	}

	// "items"
	if a.Items != nil {
		sb.WriteString(`,"items":`)
		data, err := a.Items.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("arrayNode items: %w", err)
		}
		sb.Write(data)
	}

	sb.WriteByte('}')
	return []byte(sb.String()), nil
}

//---------------------------------------------------------------------
// UnionTypeNode (anyOf)
//---------------------------------------------------------------------

func (u UnionTypeNode) TypeID() scanner.TypeID { return u.TypeID_ }

func (u UnionTypeNode) jsonSchemaMarker() {}

// Marshal as:
//
//	{
//	  "anyOf": [
//	    <ObjectNode-with-!type-const>,
//	    <ObjectNode-with-!type-const>,
//	    ...
//	  ]
//	}
func (u UnionTypeNode) MarshalJSON() ([]byte, error) {
	var sb strings.Builder
	sb.WriteString(`"{anyOf":[`)

	for i, obj := range u.Options {
		if i > 0 {
			sb.WriteByte(',')
		}

		// We'll produce a new node with a prepended property for the !type:
		//   !type: { "type":"string", "const": obj.Discriminator }
		tmpNode := prependDiscriminator(obj, u.DiscriminatorPropName)
		data, err := tmpNode.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("union option %d: %w", i, err)
		}
		sb.Write(data)
	}

	sb.WriteByte(']')
	sb.WriteByte('}')

	return []byte(sb.String()), nil
}

// prependDiscriminator returns an ObjectNode that has an extra property
// at the front: e.g. !type => { type:"string", const:"(the Discriminator)" },
// making that property required.
func prependDiscriminator(o ObjectNode, discPropName string) ObjectNode {
	if discPropName == "" {
		discPropName = DefaultDiscriminatorPropName
	}
	newProps := make([]ObjectProp, len(o.Properties)+1)
	newProps[0] = ObjectProp{
		Name: discPropName,
		Schema: PropertyNode[string]{
			Typ:   "string",
			Const: &o.Discriminator, // the type name
		},
		Optional: false, // must be required
	}
	for i, prop := range o.Properties {
		newProps[i+1] = prop
	}
	return ObjectNode{
		Desc:          o.Desc,
		Properties:    newProps,
		Discriminator: o.Discriminator,
	}
}

//---------------------------------------------------------------------
// Helper for string encoding
//---------------------------------------------------------------------

func encodeString(sb *strings.Builder, s string) {
	b, _ := json.Marshal(s) // let standard library do the escaping
	sb.Write(b)
}
