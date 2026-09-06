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

// TestInterfaceFuncTypeSchemaCallable proves that a free-function schema
// entrypoint for a registered sealed interface (also an invalid method
// receiver base, just like a named pointer type) gets a generated, callable
// implementation whose JSON schema correctly reflects the registered
// implementation, instead of either being silently dropped or generated
// with an uncompilable `func (InterfaceFuncType) ...` method signature.
func TestInterfaceFuncTypeSchemaCallable(t *testing.T) {
	raw := InterfaceFuncTypeSchema(nil)

	var got struct {
		AnyOf []struct {
			Type       string         `json:"type"`
			Properties map[string]any `json:"properties"`
			Required   []string       `json:"required"`
		} `json:"anyOf"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("InterfaceFuncTypeSchema() returned invalid JSON: %v\n%s", err, raw)
	}
	if len(got.AnyOf) != 1 {
		t.Fatalf("InterfaceFuncTypeSchema() anyOf has %d entries, want 1\n%s", len(got.AnyOf), raw)
	}
	if _, ok := got.AnyOf[0].Properties["name"]; !ok {
		t.Fatalf("InterfaceFuncTypeSchema() implementation is missing the InterfaceFuncImpl.Name property\n%s", raw)
	}
}
