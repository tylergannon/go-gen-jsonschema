package interfaces

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"go.yaml.in/yaml/v4"
)

func TestGeneratedYAMLUnmarshalSimple(t *testing.T) {
	var _ yaml.Unmarshaler = (*FancyStruct)(nil)

	var got FancyStruct
	input := []byte("iface:\n  type: TestInterface1\n  field1: one\n")
	if err := yaml.Unmarshal(input, &got); err != nil {
		t.Fatal(err)
	}
	impl, ok := got.IFace.(TestInterface1)
	if !ok || impl.Field1 != "one" {
		t.Fatalf("iface = %#v, want TestInterface1{Field1: %q}", got.IFace, "one")
	}
}

func TestLegacyInterfaceSliceDecode(t *testing.T) {
	var got FancyStruct
	input := []byte(`{"iface":{"type":"TestInterface1","field1":"one"},"ifaces":[{"type":"TestInterface2","fork3":3},{"type":"PointerToTestInterface","fork99":99}]}`)
	if err := json.Unmarshal(input, &got); err != nil {
		t.Fatal(err)
	}
	want := []TestInterface{TestInterface2{Fork3: 3}, &PointerToTestInterface{Fork99: 99}}
	if !reflect.DeepEqual(got.IFaces, want) {
		t.Fatalf("ifaces = %#v, want %#v", got.IFaces, want)
	}
}

func TestLegacyInterfaceSliceErrorIsIndexedAndTransactional(t *testing.T) {
	original := FancyStruct{IFace: TestInterface1{Field1: "original"}, IFaces: []TestInterface{TestInterface2{Fork3: 7}}}
	got := original
	input := []byte(`{"iface":{"type":"TestInterface1","field1":"replacement"},"ifaces":[{"type":"TestInterface2","fork3":3},{"type":"unknown"}]}`)
	err := json.Unmarshal(input, &got)
	if err == nil || !strings.Contains(err.Error(), "ifaces[1]") {
		t.Fatalf("error = %v, want indexed failure", err)
	}
	if !reflect.DeepEqual(got, original) {
		t.Fatalf("failed decode mutated destination: got %#v, want %#v", got, original)
	}
}
