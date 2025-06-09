package jsonschema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectSchema(t *testing.T) {
	t.Run("produces a valid JSON Schema with type:object", func(t *testing.T) {
		schema := &ObjectSchema{Description: "Test schema"}
		data, err := json.Marshal(schema)
		assert.NoError(t, err)
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		assert.NoError(t, err)
		assert.Equal(t, "object", result["type"])
	})

	t.Run("includes properties when they are provided", func(t *testing.T) {
		schema := &ObjectSchema{Description: "Test schema"}
		schema.AddProperty("foo", BoolSchema("A boolean property"))
		schema.AddProperty("bar", StringSchema("A string property"))
		data, err := json.Marshal(schema)
		assert.NoError(t, err)
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		assert.NoError(t, err)
		props, ok := result["properties"].(map[string]interface{})
		assert.True(t, ok)
		foo := props["foo"].(map[string]interface{})
		bar := props["bar"].(map[string]interface{})
		assert.Equal(t, "boolean", foo["type"])
		assert.Equal(t, "A boolean property", foo["description"])
		assert.Equal(t, "string", bar["type"])
		assert.Equal(t, "A string property", bar["description"])
	})

	t.Run("includes required properties when they are added", func(t *testing.T) {
		schema := &ObjectSchema{Description: "Test schema"}
		schema.AddProperty("foo", BoolSchema("A boolean property"))
		schema.AddRequiredProperty("bar", StringSchema("A required string property"))
		data, err := json.Marshal(schema)
		assert.NoError(t, err)
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		assert.NoError(t, err)
		required, ok := result["required"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, required, 1)
		assert.Equal(t, "bar", required[0])
	})

	t.Run("includes the description when provided", func(t *testing.T) {
		schema := &ObjectSchema{Description: "Test schema description"}
		data, err := json.Marshal(schema)
		assert.NoError(t, err)
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		assert.NoError(t, err)
		assert.Equal(t, "Test schema description", result["description"])
	})

	t.Run("includes additionalProperties when provided", func(t *testing.T) {
		schema := &ObjectSchema{Description: "Test schema", AdditionalProperties: true}
		data, err := json.Marshal(schema)
		assert.NoError(t, err)
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		assert.NoError(t, err)
		assert.Equal(t, true, result["additionalProperties"])
	})

	t.Run("strict mode", func(t *testing.T) {
		t.Run("makes all properties required and additionalProperties false", func(t *testing.T) {
			schema := &ObjectSchema{Description: "Test schema with strict mode", Strict: true}
			schema.AddProperty("foo", BoolSchema("A boolean property"))
			schema.AddProperty("bar", StringSchema("A string property"))
			data, err := json.Marshal(schema)
			assert.NoError(t, err)
			expected := []byte(`{"type":"object","description":"Test schema with strict mode","properties":{"foo":{"description":"A boolean property","type":"boolean"},"bar":{"description":"A string property","type":"string"}},"required":["foo","bar"],"additionalProperties":false}`)
			assert.Equal(t, string(expected), string(data))
			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			assert.NoError(t, err)
			assert.Equal(t, false, result["additionalProperties"])
			required, ok := result["required"].([]interface{})
			assert.True(t, ok)
			assert.Len(t, required, 2)
			assert.Contains(t, required, "foo")
			assert.Contains(t, required, "bar")
		})

		t.Run("overrides explicit additionalProperties when Strict is true", func(t *testing.T) {
			schema := &ObjectSchema{Description: "Test schema with strict mode", Strict: true, AdditionalProperties: true}
			schema.AddProperty("foo", BoolSchema("A boolean property"))
			data, err := json.Marshal(schema)
			assert.NoError(t, err)
			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			assert.NoError(t, err)
			assert.Equal(t, false, result["additionalProperties"])
		})

		t.Run("overrides explicit required properties when Strict is true", func(t *testing.T) {
			schema := &ObjectSchema{Description: "Test schema with strict mode", Strict: true, Required: []string{"foo"}}
			schema.AddProperty("foo", BoolSchema("A boolean property"))
			schema.AddProperty("bar", StringSchema("A string property"))
			data, err := json.Marshal(schema)
			assert.NoError(t, err)
			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			assert.NoError(t, err)
			required, ok := result["required"].([]interface{})
			assert.True(t, ok)
			assert.Len(t, required, 2)
			assert.Contains(t, required, "foo")
			assert.Contains(t, required, "bar")
		})
	})

	t.Run("does not include the Strict field in the marshaled output", func(t *testing.T) {
		schema := &ObjectSchema{Description: "Test schema", Strict: true}
		data, err := json.Marshal(schema)
		assert.NoError(t, err)
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		assert.NoError(t, err)
		_, hasStrict := result["Strict"]
		_, hasStrictLower := result["strict"]
		assert.False(t, hasStrict)
		assert.False(t, hasStrictLower)
	})
}
