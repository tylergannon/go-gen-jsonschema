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
	additionalPropertiesValues := []struct {
		name  string
		value any
	}{
		{name: "true", value: true},
		{name: "false", value: false},
		{name: "schema", value: StringSchema("extra values")},
	}

	for _, additionalProperties := range additionalPropertiesValues {
		t.Run(additionalProperties.name, func(t *testing.T) {
			schema := JSONSchema{
				Type: Object,
				Properties: map[string]SchemaNode{
					"zeta":  StringSchema("z"),
					"alpha": StringSchema("a"),
				},
				Required:             []string{"provided-by-caller"},
				AdditionalProperties: additionalProperties.value,
				Strict:               true,
			}

			first, err := json.Marshal(schema)
			require.NoError(t, err)
			for range 20 {
				current, err := json.Marshal(schema)
				require.NoError(t, err)
				require.Equal(t, first, current)
			}

			var document map[string]any
			require.NoError(t, json.Unmarshal(first, &document))
			require.Equal(t, false, document["additionalProperties"])
			require.Equal(t, []any{"alpha", "zeta"}, document["required"])

			compiled := compileIssue59Schema(t, schema)
			require.NoError(t, compiled.Validate(map[string]any{"alpha": "ok", "zeta": "ok"}))
			require.Error(t, compiled.Validate(map[string]any{"alpha": "ok"}))
			require.Error(t, compiled.Validate(map[string]any{"alpha": "ok", "zeta": "ok", "extra": "rejected"}))
		})
	}
}

func TestIssue59ObjectSchemaStrictOverridesSettings(t *testing.T) {
	additionalPropertiesValues := []struct {
		name  string
		value any
	}{
		{name: "true", value: true},
		{name: "false", value: false},
		{name: "schema", value: StringSchema("extra values")},
	}

	for _, additionalProperties := range additionalPropertiesValues {
		t.Run(additionalProperties.name, func(t *testing.T) {
			schema := &ObjectSchema{
				Strict:               true,
				Required:             []string{"provided-by-caller"},
				AdditionalProperties: additionalProperties.value,
			}
			schema.AddProperty("alpha", StringSchema("a"))
			schema.AddProperty("zeta", StringSchema("z"))

			data, err := json.Marshal(schema)
			require.NoError(t, err)
			var document map[string]any
			require.NoError(t, json.Unmarshal(data, &document))
			require.Equal(t, false, document["additionalProperties"])
			require.Equal(t, []any{"alpha", "zeta"}, document["required"])

			compiled := compileIssue59Schema(t, schema)
			require.NoError(t, compiled.Validate(map[string]any{"alpha": "ok", "zeta": "ok"}))
			require.Error(t, compiled.Validate(map[string]any{"alpha": "ok"}))
			require.Error(t, compiled.Validate(map[string]any{"alpha": "ok", "zeta": "ok", "extra": "rejected"}))
		})
	}
}

func TestIssue59NonStrictSettingsRemainEffective(t *testing.T) {
	additionalPropertiesValues := []struct {
		name       string
		value      any
		extraValue any
		allows     bool
	}{
		{name: "true", value: true, extraValue: 123, allows: true},
		{name: "false", value: false, extraValue: 123, allows: false},
		{name: "schema", value: StringSchema("extra values"), extraValue: "accepted", allows: true},
	}

	builders := []struct {
		name  string
		build func(any) SchemaNode
	}{
		{name: "JSONSchema", build: func(additionalProperties any) SchemaNode {
			return JSONSchema{
				Type:                 Object,
				Properties:           map[string]SchemaNode{"alpha": StringSchema("a"), "beta": StringSchema("b")},
				Required:             []string{"alpha"},
				AdditionalProperties: additionalProperties,
			}
		}},
		{name: "ObjectSchema", build: func(additionalProperties any) SchemaNode {
			schema := &ObjectSchema{AdditionalProperties: additionalProperties, Required: []string{"alpha"}}
			schema.AddProperty("alpha", StringSchema("a"))
			schema.AddProperty("beta", StringSchema("b"))
			return schema
		}},
	}

	for _, builder := range builders {
		for _, additionalProperties := range additionalPropertiesValues {
			t.Run(builder.name+"/"+additionalProperties.name, func(t *testing.T) {
				compiled := compileIssue59Schema(t, builder.build(additionalProperties.value))
				require.NoError(t, compiled.Validate(map[string]any{"alpha": "ok"}))
				extra := map[string]any{"alpha": "ok", "extra": additionalProperties.extraValue}
				if additionalProperties.allows {
					require.NoError(t, compiled.Validate(extra))
				} else {
					require.Error(t, compiled.Validate(extra))
				}
			})
		}
	}
}

func TestIssue59EmptyEnumReturnsMarshalError(t *testing.T) {
	node := EnumSchema[issue59Label]("empty")
	_, err := json.Marshal(node)
	require.Error(t, err)
	require.ErrorContains(t, err, "EnumSchema requires at least one value")
}
