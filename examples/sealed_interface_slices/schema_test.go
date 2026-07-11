package sealed_interface_slices

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestBatchSchemaPlacesUnionUnderArrayItems(t *testing.T) {
	var schema struct {
		Properties map[string]struct {
			Type  string `json:"type"`
			Items struct {
				AnyOf []struct {
					Properties map[string]struct {
						Const string `json:"const"`
					} `json:"properties"`
				} `json:"anyOf"`
			} `json:"items"`
			AnyOf []json.RawMessage `json:"anyOf"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(Batch{}.Schema(), &schema); err != nil {
		t.Fatal(err)
	}
	events := schema.Properties["events"]
	if events.Type != "array" {
		t.Fatalf("events type = %q, want array", events.Type)
	}
	if len(events.AnyOf) != 0 {
		t.Fatalf("events has property-level anyOf: %s", events.AnyOf)
	}
	if len(events.Items.AnyOf) != 2 {
		t.Fatalf("events.items.anyOf has %d options, want 2", len(events.Items.AnyOf))
	}
	wantDiscriminators := []string{"Created", "Deleted"}
	for i, want := range wantDiscriminators {
		if got := events.Items.AnyOf[i].Properties["!kind"].Const; got != want {
			t.Fatalf("events.items.anyOf[%d] !kind const = %q, want %q", i, got, want)
		}
	}
}

func TestBatchUnmarshalInterfaceSlice(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var got Batch
		if err := json.Unmarshal([]byte(`{"events":[]}`), &got); err != nil {
			t.Fatal(err)
		}
		if got.Events == nil || len(got.Events) != 0 {
			t.Fatalf("events = %#v, want non-nil empty slice", got.Events)
		}
	})

	t.Run("mixed value and pointer implementations", func(t *testing.T) {
		var got Batch
		input := []byte(`{"events":[{"!kind":"Created","name":"first"},{"!kind":"Deleted","id":"gone"}]}`)
		if err := json.Unmarshal(input, &got); err != nil {
			t.Fatal(err)
		}
		want := []Event{Created{Name: "first"}, &Deleted{ID: "gone"}}
		if !reflect.DeepEqual(got.Events, want) {
			t.Fatalf("events = %#v, want %#v", got.Events, want)
		}
		remarshaled, err := json.Marshal(got)
		if err != nil {
			t.Fatal(err)
		}
		if string(remarshaled) != `{"events":[{"name":"first"},{"id":"gone"}]}` {
			t.Fatalf("re-marshaled batch = %s", remarshaled)
		}
	})

	t.Run("missing preserves and null clears", func(t *testing.T) {
		original := Batch{Events: []Event{Created{Name: "original"}}}
		missing := original
		if err := json.Unmarshal([]byte(`{}`), &missing); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(missing, original) {
			t.Fatalf("missing events changed destination: got %#v, want %#v", missing, original)
		}

		null := original
		if err := json.Unmarshal([]byte(`{"events":null}`), &null); err != nil {
			t.Fatal(err)
		}
		if null.Events != nil {
			t.Fatalf("null events = %#v, want nil", null.Events)
		}
	})
}

func TestBatchUnmarshalErrorsAreIndexedAndTransactional(t *testing.T) {
	tests := []struct {
		name       string
		second     string
		wantDetail string
	}{
		{name: "missing discriminator", second: `{"id":"gone"}`, wantDetail: "no discriminator property '!kind' found"},
		{name: "unknown discriminator", second: `{"!kind":"Other","id":"gone"}`, wantDetail: "unknown discriminator: Other"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := Batch{Events: []Event{Created{Name: "original"}}}
			got := original
			input := `{"events":[{"!kind":"Created","name":"replacement"},` + test.second + `]}`
			err := json.Unmarshal([]byte(input), &got)
			if err == nil {
				t.Fatal("invalid discriminator unexpectedly succeeded")
			}
			if !strings.Contains(err.Error(), "events[1]") || !strings.Contains(err.Error(), test.wantDetail) {
				t.Fatalf("error = %q, want index and %q", err, test.wantDetail)
			}
			if !reflect.DeepEqual(got, original) {
				t.Fatalf("failed decode mutated destination: got %#v, want %#v", got, original)
			}
		})
	}
}
