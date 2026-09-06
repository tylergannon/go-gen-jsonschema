package entrypoints

import (
	"encoding/json"
	"testing"
)

// TestPointerFuncTypeSchemaCallable proves that a free-function schema
// entrypoint for a type whose underlying type is a pointer (so Go forbids
// declaring a method on it) actually gets a generated, callable
// implementation instead of being silently dropped from codegen.
func TestPointerFuncTypeSchemaCallable(t *testing.T) {
	raw := PointerFuncTypeSchema(nil)

	var got struct {
		Type        string `json:"type"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("PointerFuncTypeSchema() returned invalid JSON: %v\n%s", err, raw)
	}
	if got.Type != "integer" {
		t.Fatalf("PointerFuncTypeSchema() type = %q, want %q", got.Type, "integer")
	}
	if got.Description == "" {
		t.Fatalf("PointerFuncTypeSchema() description is empty, want PointerFuncType's doc comment")
	}
}
