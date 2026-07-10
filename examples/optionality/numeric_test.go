package optionality

import (
	"encoding/json"
	"testing"
)

func TestNumericWrapperSchemas(t *testing.T) {
	var schema struct {
		Properties map[string]struct {
			Type any `json:"type"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(NumericConfig{}.Schema(), &schema); err != nil {
		t.Fatal(err)
	}
	integerKinds := []string{"count", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64"}
	for _, kind := range integerKinds {
		assertNumericSchema(t, schema.Properties, kind, "integer")
	}
	for _, kind := range []string{"float32", "float64"} {
		assertNumericSchema(t, schema.Properties, kind, "number")
	}
}

func assertNumericSchema(t *testing.T, properties map[string]struct {
	Type any `json:"type"`
}, name, want string) {
	t.Helper()
	for _, field := range []string{name, "optional_" + name} {
		if got := properties[field].Type; got != want {
			t.Errorf("%s type = %#v, want %q", field, got, want)
		}
	}
	got, ok := properties["nullable_"+name].Type.([]any)
	if !ok || len(got) != 2 || got[0] != want || got[1] != "null" {
		t.Errorf("nullable_%s type = %#v, want [%q null]", name, properties["nullable_"+name].Type, want)
	}
}
