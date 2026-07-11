package interfaces

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestLegacyInterfaceSliceDecode(t *testing.T) {
	var got FancyStruct
	input := []byte(`{"iface":{"!type":"TestInterface1","field1":"one"},"ifaces":[{"!type":"TestInterface2","fork3":3},{"!type":"PointerToTestInterface","fork99":99}]}`)
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
	input := []byte(`{"iface":{"!type":"TestInterface1","field1":"replacement"},"ifaces":[{"!type":"TestInterface2","fork3":3},{"!type":"unknown"}]}`)
	err := json.Unmarshal(input, &got)
	if err == nil || !strings.Contains(err.Error(), "ifaces[1]") {
		t.Fatalf("error = %v, want indexed failure", err)
	}
	if !reflect.DeepEqual(got, original) {
		t.Fatalf("failed decode mutated destination: got %#v, want %#v", got, original)
	}
}
