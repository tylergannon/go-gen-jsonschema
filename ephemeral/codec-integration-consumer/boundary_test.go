package boundary

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"reflect"
	"testing"
)

func TestCandidateBoundary(t *testing.T) {
	original := Pair{Left: Created{Name: "left"}, Right: &Deleted{ID: 7}, Events: []Event{Created{Name: "list"}, &Deleted{ID: 9}}, Note: jsonschema.Optional[string]{Present: true, Value: ""}}
	for _, input := range []any{original, &original} {
		Calls = 0
		data, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		if Calls != 2 {
			t.Fatalf("concrete custom marshaler called %d times, expected 2", Calls)
		}
		if err = original.ValidateJSON(data); err != nil {
			t.Fatalf("encoded document rejected: %s: %v", data, err)
		}
		var decoded Pair
		if err = json.Unmarshal(data, &decoded); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(original, decoded) {
			t.Fatalf("semantic roundtrip mismatch: %#v != %#v", original, decoded)
		}
		var object map[string]json.RawMessage
		if err = json.Unmarshal(data, &object); err != nil {
			t.Fatal(err)
		}
		if _, ok := object["extra"]; ok {
			t.Fatal("absent Optional union emitted")
		}
		if string(object["note"]) != `""` {
			t.Fatalf("present empty string lost: %s", data)
		}
		t.Logf("schema-valid roundtrip: %s", data)
	}
}
