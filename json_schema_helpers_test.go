package jsonschema

import (
	"bytes"
	"encoding/json"
	"testing"

	validator "github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/require"
)

type issue59Count int
type issue59Label string

func compileIssue59Schema(t *testing.T, node SchemaNode) *validator.Schema {
	t.Helper()
	data, err := json.Marshal(node)
	require.NoError(t, err)
	doc, err := validator.UnmarshalJSON(bytes.NewReader(data))
	require.NoError(t, err)

	compiler := validator.NewCompiler()
	require.NoError(t, compiler.AddResource("issue-59.json", doc))
	schema, err := compiler.Compile("issue-59.json")
	require.NoError(t, err)
	return schema
}

func TestIssue59PrimitiveHelpersRecognizeNamedTypes(t *testing.T) {
	t.Run("named integer const", func(t *testing.T) {
		node := ConstSchema(issue59Count(3), "count")
		data, err := json.Marshal(node)
		require.NoError(t, err)
		var document map[string]any
		require.NoError(t, json.Unmarshal(data, &document))
		require.Equal(t, "integer", document["type"])
		require.Equal(t, float64(3), document["const"])

		schema := compileIssue59Schema(t, node)
		require.NoError(t, schema.Validate(3))
		require.Error(t, schema.Validate(4))
	})

	t.Run("named string const", func(t *testing.T) {
		node := ConstSchema(issue59Label("ready"), "status")
		data, err := json.Marshal(node)
		require.NoError(t, err)
		var document map[string]any
		require.NoError(t, json.Unmarshal(data, &document))
		require.Equal(t, "string", document["type"])
		require.Equal(t, "ready", document["const"])

		schema := compileIssue59Schema(t, node)
		require.NoError(t, schema.Validate("ready"))
		require.Error(t, schema.Validate("blocked"))
	})

	t.Run("named integer enum", func(t *testing.T) {
		node := EnumSchema("count", issue59Count(1), issue59Count(2))
		data, err := json.Marshal(node)
		require.NoError(t, err)
		var document map[string]any
		require.NoError(t, json.Unmarshal(data, &document))
		require.Equal(t, "integer", document["type"])
		require.Equal(t, []any{float64(1), float64(2)}, document["enum"])

		schema := compileIssue59Schema(t, node)
		require.NoError(t, schema.Validate(1))
		require.Error(t, schema.Validate(3))
	})

	t.Run("named string enum", func(t *testing.T) {
		node := EnumSchema("status", issue59Label("ready"), issue59Label("done"))
		data, err := json.Marshal(node)
		require.NoError(t, err)
		var document map[string]any
		require.NoError(t, json.Unmarshal(data, &document))
		require.Equal(t, "string", document["type"])
		require.Equal(t, []any{"ready", "done"}, document["enum"])

		schema := compileIssue59Schema(t, node)
		require.NoError(t, schema.Validate("done"))
		require.Error(t, schema.Validate("blocked"))
	})
}

func TestIssue59ObjectSchemaEscapesPropertyNames(t *testing.T) {
	propertyNames := []string{
		`quote"name`,
		`slash\name`,
		"control\n\tname",
		"unicode-é-雪",
	}
	schema := &ObjectSchema{}
	instance := make(map[string]any, len(propertyNames))
	for _, name := range propertyNames {
		schema.AddProperty(name, StringSchema("value"))
		instance[name] = "ok"
	}

	data, err := json.Marshal(schema)
	require.NoError(t, err)
	require.True(t, json.Valid(data))
	var document map[string]any
	require.NoError(t, json.Unmarshal(data, &document))
	properties, ok := document["properties"].(map[string]any)
	require.True(t, ok)
	for _, name := range propertyNames {
		require.Contains(t, properties, name)
	}

	compiled := compileIssue59Schema(t, schema)
	require.NoError(t, compiled.Validate(instance))
}

func TestIssue59JSONSchemaStrictRequiredOrderIsStable(t *testing.T) {
	schema := JSONSchema{
		Type: Object,
		Properties: map[string]SchemaNode{
			"zeta":   StringSchema("z"),
			"alpha":  StringSchema("a"),
			"middle": StringSchema("m"),
		},
		Required:             []string{"provided-by-caller"},
		AdditionalProperties: true,
		Strict:               true,
	}

	first, err := json.Marshal(schema)
	require.NoError(t, err)
	for i := 0; i < 20; i++ {
		current, err := json.Marshal(schema)
		require.NoError(t, err)
		require.Equal(t, first, current)
	}

	var document map[string]any
	require.NoError(t, json.Unmarshal(first, &document))
	require.Equal(t, false, document["additionalProperties"])
	require.Equal(t, []any{"alpha", "middle", "zeta"}, document["required"])
}

func TestIssue59EmptyEnumReturnsMarshalError(t *testing.T) {
	node := EnumSchema[issue59Label]("empty")
	_, err := json.Marshal(node)
	require.Error(t, err)
	require.ErrorContains(t, err, "EnumSchema requires at least one value")
}
