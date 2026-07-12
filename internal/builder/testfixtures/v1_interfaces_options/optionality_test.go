package v1_interfaces_options

import (
	"encoding/json"
	"testing"
)

func TestInterfaceSliceDecode(t *testing.T) {
	var got Owner
	input := []byte(`{"if":{"!kind":"impl_one","x":"required"},"ifs":[{"!kind":"Impl1","x":"one"},{"!kind":"Impl2","y":2}]}`)
	if err := json.Unmarshal(input, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.IFaces) != 2 {
		t.Fatalf("interfaces = %#v, want two values", got.IFaces)
	}
	first, firstOK := got.IFaces[0].(Impl1)
	second, secondOK := got.IFaces[1].(Impl2)
	if !firstOK || first.X != "one" || !secondOK || second.Y != 2 {
		t.Fatalf("interfaces = %#v", got.IFaces)
	}
}

func TestExplicitInterfaceWireValuesDriveSchemaAndDecode(t *testing.T) {
	var schema struct {
		Properties map[string]struct {
			AnyOf []struct {
				Properties map[string]struct {
					Const string `json:"const"`
				} `json:"properties"`
			} `json:"anyOf"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(Owner{}.Schema(), &schema); err != nil {
		t.Fatal(err)
	}
	options := schema.Properties["if"].AnyOf
	if len(options) != 2 || options[0].Properties["!kind"].Const != "impl_one" || options[1].Properties["!kind"].Const != `impl "two"` {
		t.Fatalf("wire discriminators = %#v", options)
	}

	var got Owner
	if err := json.Unmarshal([]byte(`{"if":{"!kind":"impl \"two\"","y":7},"ifs":[]}`), &got); err != nil {
		t.Fatal(err)
	}
	if impl, ok := got.IF.(Impl2); !ok || impl.Y != 7 {
		t.Fatalf("decoded IF = %#v", got.IF)
	}
	if err := json.Unmarshal([]byte(`{"if":{"!kind":"Impl2","y":7},"ifs":[]}`), &got); err == nil {
		t.Fatal("legacy type-name discriminator unexpectedly accepted for explicitly named field")
	}
}

func TestOptionalInterfaceDecodeIsTransactional(t *testing.T) {
	original := Owner{IF: Impl1{X: "original"}}
	got := original
	if err := json.Unmarshal([]byte(`{"if":{"!kind":"impl \"two\"","y":2},"optional_if":{"!kind":"unknown"}}`), &got); err == nil {
		t.Fatal("unknown optional interface discriminator unexpectedly succeeded")
	}
	if current, ok := got.IF.(Impl1); !ok || current.X != "original" || got.OptionalIF.Present {
		t.Fatalf("failed decode mutated destination: %#v", got)
	}
}

func TestOptionalInterfaceStates(t *testing.T) {
	var missing Owner
	if err := json.Unmarshal([]byte(`{"if":{"!kind":"impl_one","x":"required"}}`), &missing); err != nil {
		t.Fatal(err)
	}
	if missing.OptionalIF.Present {
		t.Fatal("missing optional interface is present")
	}

	var present Owner
	if err := json.Unmarshal([]byte(`{"if":{"!kind":"impl_one","x":"required"},"optional_if":{"!kind":"Impl2","y":0}}`), &present); err != nil {
		t.Fatal(err)
	}
	value, ok := present.OptionalIF.Value.(Impl2)
	if !present.OptionalIF.Present || !ok || value.Y != 0 {
		t.Fatalf("present optional interface = %#v", present.OptionalIF)
	}

	got := Owner{IF: Impl1{X: "original"}}
	if err := json.Unmarshal([]byte(`{"if":{"!kind":"impl_one","x":"required"},"optional_if":null}`), &got); err == nil {
		t.Fatal("null optional interface unexpectedly succeeded")
	}
	if current, ok := got.IF.(Impl1); !ok || current.X != "original" || got.OptionalIF.Present {
		t.Fatalf("null decode mutated destination: %#v", got)
	}
}
