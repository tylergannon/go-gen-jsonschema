package boundary

import (
	"bytes"
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"reflect"
	"testing"
)

func TestFieldEnumBoundary(t *testing.T) {
	event := Pair{Left: Created{Name: "first"}, Right: &Deleted{ID: 2}, Events: []Event{}}
	for _, present := range []bool{false, true} {
		original := Config{Event: event, Direct: Quiet, Numeric: Loud, Optional: jsonschema.Optional[Mode]{Present: present}, Nullable: jsonschema.Nullable[Mode]{Present: present}}
		if present {
			original.Optional.Value = Loud
			original.Nullable.Value = Quiet
		}
		for _, input := range []any{original, &original} {
			data, err := json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
			if err = original.ValidateJSON(data); err != nil {
				t.Fatalf("invalid generated JSON %s: %v", data, err)
			}
			var decoded Config
			if err = json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(original, decoded) {
				t.Fatalf("roundtrip mismatch: %#v != %#v", original, decoded)
			}
			var wire map[string]json.RawMessage
			if err = json.Unmarshal(data, &wire); err != nil {
				t.Fatal(err)
			}
			if string(wire["direct"]) != `"Quiet"` || string(wire["numeric"]) != "7" {
				t.Fatalf("wrong field mode: %s", data)
			}
			if !present && (wire["optional"] != nil || string(wire["nullable"]) != "null") {
				t.Fatalf("wrong absence/null: %s", data)
			}
			for _, invalid := range [][]byte{[]byte(`"Missing"`), []byte(`0`), []byte(`null`)} {
				bad := bytes.Replace(data, []byte(`"direct":"Quiet"`), append([]byte(`"direct":`), invalid...), 1)
				unchanged := original
				if err = json.Unmarshal(bad, &unchanged); err == nil {
					t.Fatalf("accepted invalid enum %s", bad)
				}
				if !reflect.DeepEqual(unchanged, original) {
					t.Fatal("decode failure mutated owner")
				}
			}
			t.Logf("schema-valid field-specific roundtrip: %s", data)
		}
	}
	invalid := Config{Event: event, Direct: 0, Numeric: 7}
	if _, err := json.Marshal(invalid); err == nil {
		t.Fatal("accepted undeclared zero")
	}
}
