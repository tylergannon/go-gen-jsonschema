package ref_types

import (
	"encoding/json"
	"fmt"
	"testing"
)

const (
	openAIStructuredOutputsSupportedSchemas = "https://developers.openai.com/api/docs/guides/structured-outputs#supported-schemas"
	openAIFunctionCallingStrictMode         = "https://developers.openai.com/api/docs/guides/function-calling#strict-mode"
)

// TestNullableConfigMatchesOpenAIStrictSchemaRules checks the documented
// strict-schema invariants without requiring network access or an API key.
//
// OpenAI documents that every object must set additionalProperties to false,
// every property must be required, and null is the supported representation of
// an optional value in strict mode. The supported-schema reference lists enum,
// anyOf, and definitions as supported constructs.
func TestNullableConfigMatchesOpenAIStrictSchemaRules(t *testing.T) {
	var schema map[string]any
	if err := json.Unmarshal(NullableConfig{}.Schema(), &schema); err != nil {
		t.Fatal(err)
	}
	if schema["type"] != "object" {
		t.Fatalf("root type = %v, want object; see %s", schema["type"], openAIStructuredOutputsSupportedSchemas)
	}
	if _, ok := schema["anyOf"]; ok {
		t.Fatalf("root must not be anyOf; see %s", openAIStructuredOutputsSupportedSchemas)
	}
	assertOpenAIStrictSchemaNode(t, schema, "#")
}

func assertOpenAIStrictSchemaNode(t *testing.T, node any, path string) {
	t.Helper()

	switch value := node.(type) {
	case map[string]any:
		if value["type"] == "object" {
			if additional, ok := value["additionalProperties"].(bool); !ok || additional {
				t.Fatalf("%s: object must set additionalProperties:false; see %s", path, openAIFunctionCallingStrictMode)
			}
			properties, _ := value["properties"].(map[string]any)
			requiredList, _ := value["required"].([]any)
			required := make(map[string]bool, len(requiredList))
			for _, item := range requiredList {
				name, ok := item.(string)
				if !ok {
					t.Fatalf("%s: required entry %v is not a string", path, item)
				}
				required[name] = true
			}
			for name := range properties {
				if !required[name] {
					t.Fatalf("%s: property %q is not required; see %s", path, name, openAIFunctionCallingStrictMode)
				}
			}
			if len(required) != len(properties) {
				t.Fatalf("%s: required=%v does not exactly match properties; see %s", path, required, openAIFunctionCallingStrictMode)
			}
		}
		for key, child := range value {
			assertOpenAIStrictSchemaNode(t, child, fmt.Sprintf("%s/%s", path, key))
		}
	case []any:
		for index, child := range value {
			assertOpenAIStrictSchemaNode(t, child, fmt.Sprintf("%s/%d", path, index))
		}
	}
}

// TestNullableConfigAcceptedByOpenAIStructuredOutputs is the live acceptance
// check retained as an explicit reminder of the external contract. The
// deterministic strict-subset and JSON Schema behavior are covered above and
// in ref_types_test.go.
func TestNullableConfigAcceptedByOpenAIStructuredOutputs(t *testing.T) {
	t.Skip("live OpenAI Structured Outputs acceptance requires network access and an API key")
}
