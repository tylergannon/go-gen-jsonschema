package jsonschema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParentSchema_MarshalJSON(t *testing.T) {
	testData := []struct {
		item     ParentSchema
		expected string
	}{{
		item: ParentSchema{
			ObjectSchema: &ObjectSchema{
				Strict: true,
				Properties: []SchemaProperty{
					{
						Key:   "Foo",
						Value: RefSchemaEl("$defs/FooBarBaz"),
					},
				},
			},
			DefinitionsKeyName: "$defs",
			Definitions: []SchemaProperty{
				{
					Key:   "FooBarBaz",
					Value: EnumSchema("Foo Bar Baz!!!", "Nice", "Awesome", "Excellent"),
				},
			},
		},
		expected: `{"type":"object","properties":{"Foo":{"$ref":"$defs/FooBarBaz"}},"required":["Foo"],"additionalProperties":false,"$defs":{"FooBarBaz":{"description":"Foo Bar Baz!!!","enum":["Nice","Awesome","Excellent"],"type":"string"}}}`,
	}}

	for _, data := range testData {
		marshaled, err := json.Marshal(data.item)
		assert.NoError(t, err)
		assert.JSONEq(t, data.expected, string(marshaled))
	}
}

func TestObjectSchema_MarshalJSON_Empty(t *testing.T) {
	schema := &ObjectSchema{}
	bytes, err := json.Marshal(schema)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"type":"object"}`, string(bytes))
}

func TestObjectSchema_MarshalJSON_WithDescription(t *testing.T) {
	schema := &ObjectSchema{
		Description: "A test schema",
	}
	bytes, err := json.Marshal(schema)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"type":"object","description":"A test schema"}`, string(bytes))
}

func TestObjectSchema_MarshalJSON_WithProperties(t *testing.T) {
	schema := &ObjectSchema{}
	schema.AddProperty("name", StringSchema("User name"))
	schema.AddProperty("age", IntSchema("User age"))

	bytes, err := json.Marshal(schema)
	assert.NoError(t, err)
	assert.JSONEq(t, `{
		"type": "object",
		"properties": {
			"name": {"type": "string", "description": "User name"},
			"age": {"type": "integer", "description": "User age"}
		}
	}`, string(bytes))
}

func TestObjectSchema_MarshalJSON_WithRequiredProperties(t *testing.T) {
	schema := &ObjectSchema{}
	schema.AddRequiredProperty("name", StringSchema("User name"))
	schema.AddProperty("age", IntSchema("User age"))

	bytes, err := json.Marshal(schema)
	assert.NoError(t, err)
	assert.JSONEq(t, `{
		"type": "object",
		"properties": {
			"name": {"type": "string", "description": "User name"},
			"age": {"type": "integer", "description": "User age"}
		},
		"required": ["name"]
	}`, string(bytes))
}

func TestObjectSchema_MarshalJSON_WithStrictMode(t *testing.T) {
	schema := &ObjectSchema{
		Strict: true,
	}
	schema.AddProperty("name", StringSchema("User name"))
	schema.AddProperty("age", IntSchema("User age"))

	bytes, err := json.Marshal(schema)
	assert.NoError(t, err)
	assert.JSONEq(t, `{
		"type": "object",
		"properties": {
			"name": {"type": "string", "description": "User name"},
			"age": {"type": "integer", "description": "User age"}
		},
		"required": ["name", "age"],
		"additionalProperties": false
	}`, string(bytes))
}

func TestObjectSchema_MarshalJSON_WithAdditionalProperties(t *testing.T) {
	schema := &ObjectSchema{
		AdditionalProperties: true,
	}
	schema.AddProperty("name", StringSchema("User name"))

	bytes, err := json.Marshal(schema)
	assert.NoError(t, err)
	assert.JSONEq(t, `{
		"type": "object",
		"properties": {
			"name": {"type": "string", "description": "User name"}
		},
		"additionalProperties": true
	}`, string(bytes))
}

func TestObjectSchema_MarshalJSON_WithAdditionalPropertiesSchema(t *testing.T) {
	schema := &ObjectSchema{
		AdditionalProperties: StringSchema("Additional string property"),
	}
	schema.AddProperty("name", StringSchema("User name"))

	bytes, err := json.Marshal(schema)
	assert.NoError(t, err)
	assert.JSONEq(t, `{
		"type": "object",
		"properties": {
			"name": {"type": "string", "description": "User name"}
		},
		"additionalProperties": {"type": "string", "description": "Additional string property"}
	}`, string(bytes))
}

func TestObjectSchema_MarshalJSON_ComplexSchema(t *testing.T) {
	addressSchema := &ObjectSchema{}
	addressSchema.AddProperty("street", StringSchema("Street name"))
	addressSchema.AddProperty("city", StringSchema("City name"))
	addressSchema.AddProperty("zipCode", StringSchema("ZIP code"))

	userSchema := &ObjectSchema{
		Description: "User object schema",
		Strict:      true,
	}
	userSchema.AddProperty("id", IntSchema("User ID"))
	userSchema.AddProperty("name", StringSchema("User name"))
	userSchema.AddProperty("email", StringSchema("User email"))
	userSchema.AddProperty("address", addressSchema)
	userSchema.AddProperty("tags", ArraySchema(StringSchema("Tag"), "User tags"))

	bytes, err := json.Marshal(userSchema)
	assert.NoError(t, err)

	expectedJSON := `{
		"type": "object",
		"description": "User object schema",
		"properties": {
			"id": {"type": "integer", "description": "User ID"},
			"name": {"type": "string", "description": "User name"},
			"email": {"type": "string", "description": "User email"},
			"address": {
				"type": "object",
				"properties": {
					"street": {"type": "string", "description": "Street name"},
					"city": {"type": "string", "description": "City name"},
					"zipCode": {"type": "string", "description": "ZIP code"}
				}
			},
			"tags": {
				"type": "array",
				"items": {"type": "string", "description": "Tag"},
				"description": "User tags"
			}
		},
		"required": ["id", "name", "email", "address", "tags"],
		"additionalProperties": false
	}`

	assert.JSONEq(t, expectedJSON, string(bytes))
}

func TestJSONSchema_MarshalJSON(t *testing.T) {
	schema := JSONSchema{
		Type:        Object,
		Description: "Test JSON Schema",
		Properties: map[string]json.Marshaler{
			"name": StringSchema("User name"),
			"age":  IntSchema("User age"),
		},
		Required: []string{"name"},
	}

	bytes, err := json.Marshal(schema)
	assert.NoError(t, err)
	assert.JSONEq(t, `{
		"type": "object",
		"description": "Test JSON Schema",
		"properties": {
			"name": {"type": "string", "description": "User name"},
			"age": {"type": "integer", "description": "User age"}
		},
		"required": ["name"]
	}`, string(bytes))
}

func TestJSONUnionType_MarshalJSON(t *testing.T) {
	unionSchema := JSONUnionType{
		&JSONSchema{Type: String, Description: "String option"},
		&JSONSchema{Type: Integer, Description: "Integer option"},
	}

	bytes, err := json.Marshal(unionSchema)
	assert.NoError(t, err)
	assert.JSONEq(t, `{
		"anyOf": [
			{"type": "string", "description": "String option"},
			{"type": "integer", "description": "Integer option"}
		]
	}`, string(bytes))
}
