//go:build !jsonschema

package consumer

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func TestGeneratedMixedCodecMatchesSchemaAndRoundTrips(t *testing.T) {
	input := validEnvelope()
	valueJSON, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	pointerJSON, err := json.Marshal(&input)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(valueJSON, pointerJSON) {
		t.Fatalf("value and pointer encoding differ:\nvalue:   %s\npointer: %s", valueJSON, pointerJSON)
	}
	t.Logf("encoded_json=%s", valueJSON)

	var wire map[string]json.RawMessage
	if err := json.Unmarshal(valueJSON, &wire); err != nil {
		t.Fatal(err)
	}
	assertObjectString(t, wire["primary"], "kind", "created")
	assertObjectString(t, wire["alternate"], "!kind", "alt-created")
	assertObjectString(t, wire["maybe"], "kind", "deleted")
	assertUnionSlice(t, wire["events"], []string{"created", "deleted"})
	assertJSONString(t, wire["string_state"], "StateDone")
	assertJSONNumber(t, wire["numeric_state"], "2")
	assertJSONString(t, wire["optional_state"], "StateNew")
	assertJSONString(t, wire["nullable_state"], "StateDone")
	if string(wire["null_state"]) != "null" {
		t.Fatalf("null_state = %s, want null", wire["null_state"])
	}
	assertJSONString(t, wire["ordinary-name"], "ordinary")
	if _, ok := wire["Ignored"]; ok {
		t.Fatalf("ignored Go field leaked into JSON: %s", valueJSON)
	}
	var meta map[string]json.RawMessage
	if err := json.Unmarshal(wire["meta"], &meta); err != nil {
		t.Fatal(err)
	}
	if _, ok := meta["Hidden"]; ok {
		t.Fatalf("ignored nested field leaked into JSON: %s", wire["meta"])
	}

	if err := (Envelope{}).ValidateJSON(valueJSON); err != nil {
		t.Fatalf("generated encoding failed generated schema validation: %v\n%s", err, valueJSON)
	}

	got := sentinelEnvelope()
	if err := json.Unmarshal(valueJSON, &got); err != nil {
		t.Fatal(err)
	}
	assertEnvelopeMeaning(t, got)
	reencoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(valueJSON, reencoded) {
		t.Fatalf("decode/encode changed wire meaning:\nfirst:  %s\nsecond: %s", valueJSON, reencoded)
	}
}

func TestGeneratedDecodeFailuresDoNotMutateReceiver(t *testing.T) {
	validJSON, err := json.Marshal(validEnvelope())
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		edit func(map[string]any)
		want string
	}{
		{
			name: "unknown string enum",
			edit: func(wire map[string]any) { wire["string_state"] = "NotRegistered" },
			want: "unknown string-mode enum State name",
		},
		{
			name: "unknown second union discriminator",
			edit: func(wire map[string]any) {
				wire["events"].([]any)[1].(map[string]any)["kind"] = "unknown"
			},
			want: "unknown discriminator",
		},
		{
			name: "null optional enum",
			edit: func(wire map[string]any) { wire["optional_state"] = nil },
			want: "Optional value cannot be JSON null",
		},
		{
			name: "missing nullable enum",
			edit: func(wire map[string]any) { delete(wire, "nullable_state") },
			want: "required nullable string-mode enum is missing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var wire map[string]any
			if err := json.Unmarshal(validJSON, &wire); err != nil {
				t.Fatal(err)
			}
			test.edit(wire)
			invalid, err := json.Marshal(wire)
			if err != nil {
				t.Fatal(err)
			}
			got := sentinelEnvelope()
			want := sentinelEnvelope()
			err = json.Unmarshal(invalid, &got)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("failed decode mutated receiver:\ngot:  %#v\nwant: %#v", got, want)
			}
			t.Logf("rejected=%q receiver_unchanged=true", err)
		})
	}
}

func TestGeneratedValidationRejectsUnknownNumericEnum(t *testing.T) {
	validJSON, err := json.Marshal(validEnvelope())
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(validJSON, &wire); err != nil {
		t.Fatal(err)
	}
	wire["numeric_state"] = 99
	invalid, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	if err := (Envelope{}).ValidateJSON(invalid); err == nil {
		t.Fatalf("schema validation accepted undeclared numeric enum: %s", invalid)
	} else {
		t.Logf("rejected_numeric_enum=%q", err)
	}
}

func TestGeneratedCodecUsesOneParentWrapperAndSealedSwitches(t *testing.T) {
	source, err := os.ReadFile("jsonschema_gen.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, required := range []string{
		"type Alias Envelope",
		"type Wrapper struct {",
		"Alias",
		"case Created:",
		"case *Deleted:",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("generated codec lacks %q", required)
		}
	}
	if strings.Count(text, "func (e Envelope) MarshalJSON()") != 1 || strings.Count(text, "func (e *Envelope) UnmarshalJSON(data []byte)") != 1 {
		t.Fatalf("generated owner must have exactly one marshal and one unmarshal method")
	}
	for _, field := range []string{"Primary", "Alternate", "Maybe", "Events", "StringState", "OptionalState", "NullableState", "NullState"} {
		pattern := regexp.MustCompile(`(?m)^\s*` + field + `\s+json\.RawMessage\s+`)
		if !pattern.MatchString(text) {
			t.Fatalf("special field %s is not shadowed by json.RawMessage", field)
		}
	}
	for _, field := range []string{"NumericState", "Label", "Meta", "Ignored"} {
		pattern := regexp.MustCompile(`(?m)^\s*` + field + `\s+json\.RawMessage\s+`)
		if pattern.MatchString(text) {
			t.Fatalf("ordinary field %s was unnecessarily shadowed", field)
		}
	}
	if strings.Contains(text, "func (s State) MarshalJSON()") || strings.Contains(text, "func (s *State) UnmarshalJSON(") {
		t.Fatal("field-specific enum mapping leaked into a global State codec")
	}
	t.Log("owner_methods=1/1 special_fields_rawmessage=8 ordinary_fields_via_alias=4 sealed_variants=Created,*Deleted global_state_codec=false")
}

func validEnvelope() Envelope {
	return Envelope{
		Primary:       Created{Name: "first"},
		Alternate:     Created{Name: "alternate"},
		Maybe:         jsonschema.Optional[Event]{Present: true, Value: &Deleted{ID: "optional"}},
		Events:        []Event{Created{Name: "slice-value"}, &Deleted{ID: "slice-pointer"}},
		StringState:   StateDone,
		NumericState:  StateDone,
		OptionalState: jsonschema.Optional[State]{Present: true, Value: StateNew},
		NullableState: jsonschema.Nullable[State]{Present: true, Value: StateDone},
		NullState:     jsonschema.Nullable[State]{},
		Label:         "ordinary",
		Meta:          Metadata{Visible: "visible", Hidden: "secret"},
		Ignored:       "secret",
	}
}

func sentinelEnvelope() Envelope {
	return Envelope{
		Primary:       &Deleted{ID: "sentinel-primary"},
		Alternate:     &Deleted{ID: "sentinel-alternate"},
		Maybe:         jsonschema.Optional[Event]{Present: true, Value: Created{Name: "sentinel-maybe"}},
		Events:        []Event{Created{Name: "sentinel-event"}},
		StringState:   StateNew,
		NumericState:  StateNew,
		OptionalState: jsonschema.Optional[State]{Present: true, Value: StateDone},
		NullableState: jsonschema.Nullable[State]{Present: true, Value: StateNew},
		NullState:     jsonschema.Nullable[State]{Present: true, Value: StateNew},
		Label:         "sentinel",
		Meta:          Metadata{Visible: "sentinel-visible", Hidden: "sentinel-hidden"},
		Ignored:       "sentinel-ignored",
	}
}

func assertEnvelopeMeaning(t *testing.T, got Envelope) {
	t.Helper()
	if value, ok := got.Primary.(Created); !ok || value.Name != "first" {
		t.Fatalf("primary = %#v", got.Primary)
	}
	if value, ok := got.Alternate.(Created); !ok || value.Name != "alternate" {
		t.Fatalf("alternate = %#v", got.Alternate)
	}
	if value, ok := got.Maybe.Value.(*Deleted); !got.Maybe.Present || !ok || value.ID != "optional" {
		t.Fatalf("maybe = %#v", got.Maybe)
	}
	if len(got.Events) != 2 {
		t.Fatalf("events = %#v", got.Events)
	}
	if value, ok := got.Events[0].(Created); !ok || value.Name != "slice-value" {
		t.Fatalf("events[0] = %#v", got.Events[0])
	}
	if value, ok := got.Events[1].(*Deleted); !ok || value.ID != "slice-pointer" {
		t.Fatalf("events[1] = %#v", got.Events[1])
	}
	if got.StringState != StateDone || got.NumericState != StateDone || !got.OptionalState.Present || got.OptionalState.Value != StateNew || !got.NullableState.Present || got.NullableState.Value != StateDone || got.NullState.Present {
		t.Fatalf("enum fields = %#v", got)
	}
	if got.Label != "ordinary" || got.Meta.Visible != "visible" || got.Meta.Hidden != "" || got.Ignored != "" {
		t.Fatalf("ordinary fields = %#v", got)
	}
}

func assertObjectString(t *testing.T, data json.RawMessage, key, want string) {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		t.Fatal(err)
	}
	assertJSONString(t, object[key], want)
}

func assertUnionSlice(t *testing.T, data json.RawMessage, kinds []string) {
	t.Helper()
	var values []map[string]json.RawMessage
	if err := json.Unmarshal(data, &values); err != nil {
		t.Fatal(err)
	}
	if len(values) != len(kinds) {
		t.Fatalf("union slice length = %d, want %d", len(values), len(kinds))
	}
	for i, kind := range kinds {
		assertJSONString(t, values[i]["kind"], kind)
	}
}

func assertJSONString(t *testing.T, data json.RawMessage, want string) {
	t.Helper()
	var got string
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("decode JSON string from %s: %v", data, err)
	}
	if got != want {
		t.Fatalf("JSON string = %q, want %q", got, want)
	}
}

func assertJSONNumber(t *testing.T, data json.RawMessage, want string) {
	t.Helper()
	if string(data) != want {
		t.Fatalf("JSON number = %s, want %s", data, want)
	}
}
