package ref_types

import (
	"encoding/json"
	"testing"
)

func TestContainerSchemaUsesRefIntoDefs(t *testing.T) {
	var schema struct {
		Defs       map[string]json.RawMessage `json:"$defs"`
		Properties map[string]struct {
			Ref   string `json:"$ref"`
			Type  string `json:"type"`
			Items struct {
				Ref string `json:"$ref"`
			} `json:"items"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(Container{}.Schema(), &schema); err != nil {
		t.Fatal(err)
	}

	if len(schema.Defs) != 1 {
		t.Fatalf("$defs has %d entries, want 1: %v", len(schema.Defs), schema.Defs)
	}
	if _, ok := schema.Defs["Shared"]; !ok {
		t.Fatalf("$defs missing Shared entry: %v", schema.Defs)
	}

	primary, ok := schema.Properties["primary"]
	if !ok {
		t.Fatal("schema missing properties.primary")
	}
	if primary.Ref != "#/$defs/Shared" {
		t.Fatalf("properties.primary.$ref = %q, want #/$defs/Shared", primary.Ref)
	}

	others, ok := schema.Properties["others"]
	if !ok {
		t.Fatal("schema missing properties.others")
	}
	if others.Type != "array" {
		t.Fatalf("properties.others.type = %q, want array", others.Type)
	}
	if others.Items.Ref != "#/$defs/Shared" {
		t.Fatalf("properties.others.items.$ref = %q, want #/$defs/Shared", others.Items.Ref)
	}
}

func TestContainerDefsSharedMatchesTopLevelSharedSchema(t *testing.T) {
	var schema struct {
		Defs map[string]json.RawMessage `json:"$defs"`
	}
	if err := json.Unmarshal(Container{}.Schema(), &schema); err != nil {
		t.Fatal(err)
	}

	// Canonicalize both sides through json.Marshal(map[string]any) so that
	// pretty-printing/indentation differences between the standalone
	// Shared.json file and the nested $defs.Shared entry don't matter.
	var fromDefs, fromTopLevel any
	if err := json.Unmarshal(schema.Defs["Shared"], &fromDefs); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(Shared{}.Schema(), &fromTopLevel); err != nil {
		t.Fatal(err)
	}

	defsBytes, err := json.Marshal(fromDefs)
	if err != nil {
		t.Fatal(err)
	}
	topLevelBytes, err := json.Marshal(fromTopLevel)
	if err != nil {
		t.Fatal(err)
	}
	if string(defsBytes) != string(topLevelBytes) {
		t.Fatalf("$defs.Shared != Shared.json:\n  $defs.Shared: %s\n  Shared.json:  %s", defsBytes, topLevelBytes)
	}
}

func TestContainerValidateJSONThroughRef(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		input := []byte(`{"primary":{"name":"a"},"others":[{"name":"b"},{"name":"c"}]}`)
		if err := (Container{}).ValidateJSON(input); err != nil {
			t.Fatalf("expected valid input to pass: %v", err)
		}
	})

	t.Run("missing required field on directly-referenced Shared", func(t *testing.T) {
		input := []byte(`{"primary":{},"others":[]}`)
		if err := (Container{}).ValidateJSON(input); err == nil {
			t.Fatal("expected validation error for missing primary.name")
		}
	})

	t.Run("missing required field inside sliced Shared", func(t *testing.T) {
		input := []byte(`{"primary":{"name":"a"},"others":[{}]}`)
		if err := (Container{}).ValidateJSON(input); err == nil {
			t.Fatal("expected validation error for missing others[0].name")
		}
	})

	t.Run("additional property rejected via ref", func(t *testing.T) {
		input := []byte(`{"primary":{"name":"a","extra":true},"others":[]}`)
		if err := (Container{}).ValidateJSON(input); err == nil {
			t.Fatal("expected validation error for additional property on primary")
		}
	})
}

func TestSharedValidateJSON(t *testing.T) {
	if err := (Shared{}).ValidateJSON([]byte(`{"name":"a"}`)); err != nil {
		t.Fatalf("expected valid input to pass: %v", err)
	}
	if err := (Shared{}).ValidateJSON([]byte(`{}`)); err == nil {
		t.Fatal("expected validation error for missing name")
	}
}

func TestContainerRoundTrip(t *testing.T) {
	original := Container{
		Primary: Shared{Name: "primary-value"},
		Others:  []Shared{{Name: "one"}, {Name: "two"}},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	var got Container
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}

	if got.Primary != original.Primary {
		t.Fatalf("Primary = %#v, want %#v", got.Primary, original.Primary)
	}
	if len(got.Others) != len(original.Others) {
		t.Fatalf("Others = %#v, want %#v", got.Others, original.Others)
	}
	for i := range got.Others {
		if got.Others[i] != original.Others[i] {
			t.Fatalf("Others[%d] = %#v, want %#v", i, got.Others[i], original.Others[i])
		}
	}

	if err := (Container{}).ValidateJSON(data); err != nil {
		t.Fatalf("round-tripped JSON failed validation: %v", err)
	}
}
