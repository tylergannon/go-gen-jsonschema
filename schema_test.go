package jsonschema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONSchemaStrict(t *testing.T) {
	schema := JSONSchema{
		Type:        Object,
		Description: "Test schema",
		Properties: map[string]json.Marshaler{
			"foo": BoolSchema("bar"),
			"bax": StringSchema("quux"),
		},
		Strict: true,
	}

	t.Run("Sets additionalProperties to false when Strict is true", func(t *testing.T) {
		data, err := json.Marshal(schema)
		assert.NoError(t, err)
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		assert.NoError(t, err)
		assert.Equal(t, false, result["additionalProperties"])
	})

	t.Run("Includes all property keys in the required array when Strict is true", func(t *testing.T) {
		data, err := json.Marshal(schema)
		assert.NoError(t, err)
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		assert.NoError(t, err)
		required, ok := result["required"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, required, 2)
		assert.Contains(t, required, "foo")
		assert.Contains(t, required, "bax")
	})

	t.Run("Does not include the Strict field in the marshaled output", func(t *testing.T) {
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
