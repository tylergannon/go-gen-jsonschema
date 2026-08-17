package v1_interfaces_options

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	yaml "go.yaml.in/yaml/v4"
)

func TestGeneratedYAMLUnmarshalComprehensive(t *testing.T) {
	input := []byte(`
defaults: &defaults
  if:
    "!kind": impl_one
    x: merged
<<: *defaults
ifs:
  - "!kind": Impl1
    x: one
  - "!kind": Impl2
    y: 2
optional_if:
  "!kind": Impl2
  y: 3
label: ""
timeout: 0
`)
	var got Owner
	if err := yaml.Load(input, &got, yaml.WithV4Defaults()); err != nil {
		t.Fatal(err)
	}
	if len(got.IFaces) != 2 {
		t.Fatalf("interfaces = %#v, want two values", got.IFaces)
	}
	required, requiredOK := got.IF.(Impl1)
	first, firstOK := got.IFaces[0].(Impl1)
	second, secondOK := got.IFaces[1].(Impl2)
	optional, optionalOK := got.OptionalIF.Value.(Impl2)
	if !requiredOK || required.X != "json:merged" ||
		!firstOK || first.X != "json:one" || !secondOK || second.Y != 2 ||
		!got.OptionalIF.Present || !optionalOK || optional.Y != 3 ||
		!got.Label.Present || got.Label.Value != "" ||
		!got.Timeout.Present || got.Timeout.Value != 0 {
		t.Fatalf("decoded owner = %#v", got)
	}

	var nullish Owner
	if err := yaml.Load([]byte(`
if:
  "!kind": impl_one
  x: required
ifs: []
timeout: null
`), &nullish, yaml.WithV4Defaults()); err != nil {
		t.Fatal(err)
	}
	if nullish.Label.Present || nullish.Timeout.Present {
		t.Fatalf("absent optional and null nullable = %#v", nullish)
	}

	original := Owner{
		IF:      Impl1{X: "original"},
		IFaces:  []IFace{Impl2{Y: 7}},
		Label:   jsonschema.Optional[string]{Present: true, Value: "original"},
		Timeout: jsonschema.Nullable[int]{Present: true, Value: 9},
	}
	got = original
	bad := []byte(`
if:
  "!kind": impl_one
  x: replacement
ifs:
  - "!kind": Impl2
    y: 2
  - "!kind": unknown
`)
	err := yaml.Load(bad, &got, yaml.WithV4Defaults())
	if err == nil || !strings.Contains(err.Error(), "ifs[1]") {
		t.Fatalf("error = %v, want indexed ifs failure", err)
	}
	if !reflect.DeepEqual(got, original) {
		t.Fatalf("failed decode mutated destination: got %#v, want %#v", got, original)
	}

	got = original
	err = yaml.Load([]byte(`
if:
  "!kind": impl_one
  x: replacement
ifs: []
label: null
timeout: null
`), &got, yaml.WithV4Defaults())
	if err == nil || !strings.Contains(err.Error(), "Optional value cannot be JSON null") {
		t.Fatalf("error = %v, want null Optional failure", err)
	}
	if !reflect.DeepEqual(got, original) {
		t.Fatalf("failed Optional decode mutated destination: got %#v, want %#v", got, original)
	}
}

func TestGeneratedYAMLValidationUsesJSONSchema(t *testing.T) {
	valid := []byte(`
if:
  "!kind": impl_one
  x: required
ifs: []
timeout: null
`)
	if err := (Owner{}).ValidateYAML(valid); err != nil {
		t.Fatalf("valid YAML rejected: %v", err)
	}

	unknown := append(valid, []byte("surprise: true\n")...)
	if err := (Owner{}).ValidateYAML(unknown); err == nil || !strings.Contains(err.Error(), "surprise") {
		t.Fatalf("unknown-property error = %v, want surprise", err)
	}

	yamlNames := []byte(`
yaml_if:
  "!kind": impl_one
  x: required
yaml_ifs: []
timeout: null
`)
	if err := (Owner{}).ValidateYAML(yamlNames); err == nil || !strings.Contains(err.Error(), "if") {
		t.Fatalf("yaml-tag property error = %v, want schema property failure", err)
	}

	nullOptional := []byte(`
if:
  "!kind": impl_one
  x: required
ifs: []
label: null
timeout: null
`)
	if err := (Owner{}).ValidateYAML(nullOptional); err == nil || !strings.Contains(err.Error(), "label") {
		t.Fatalf("null Optional validation error = %v, want label", err)
	}
}

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
	if !firstOK || first.X != "json:one" || !secondOK || second.Y != 2 {
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
