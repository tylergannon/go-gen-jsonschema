package providers_rendering

import (
	"encoding/json"
	"testing"
)

func TestRenderedSchema(t *testing.T) {
	example := Example{
		A: "test",
		B: 42,
		C: true,
	}

	// Test RenderedSchema
	schema, err := example.RenderedSchema()
	if err != nil {
		t.Fatalf("RenderedSchema() error = %v", err)
	}

	// Verify it's valid JSON
	var schemaObj map[string]interface{}
	if err := json.Unmarshal(schema, &schemaObj); err != nil {
		t.Fatalf("Failed to parse schema JSON: %v", err)
	}

	// Check the schema structure
	props, ok := schemaObj["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema missing properties field")
	}

	// Check field A
	if aSchema, ok := props["a"].(map[string]interface{}); !ok {
		t.Fatal("Missing schema for field 'a'")
	} else if aSchema["type"] != "string" {
		t.Errorf("Field 'a' type = %v, want string", aSchema["type"])
	}

	// Check field B
	if bSchema, ok := props["b"].(map[string]interface{}); !ok {
		t.Fatal("Missing schema for field 'b'")
	} else if bSchema["type"] != "integer" {
		t.Errorf("Field 'b' type = %v, want integer", bSchema["type"])
	}

	// Check field C
	if cSchema, ok := props["c"].(map[string]interface{}); !ok {
		t.Fatal("Missing schema for field 'c'")
	} else if cSchema["type"] != "boolean" {
		t.Errorf("Field 'c' type = %v, want boolean", cSchema["type"])
	}

	t.Logf("Generated schema: %s", schema)
}

func TestStaticSchema(t *testing.T) {
	example := Example{}

	// Test the static Schema() method
	schema := example.Schema()

	// It should be a template, not rendered JSON
	schemaStr := string(schema)
	if schemaStr == "" {
		t.Fatal("Static schema is empty")
	}

	// Should contain template placeholders
	expectedPlaceholders := []string{"{{.a}}", "{{.b}}", "{{.c}}"}
	for _, placeholder := range expectedPlaceholders {
		if !contains(schemaStr, placeholder) {
			t.Errorf("Static schema missing placeholder %s", placeholder)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}
