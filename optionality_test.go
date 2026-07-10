package jsonschema

import (
	"encoding/json"
	"math"
	"testing"
)

func TestOptionalJSONStates(t *testing.T) {
	type config struct {
		Retries Optional[int] `json:"retries,omitzero"`
	}

	absent, err := json.Marshal(config{})
	if err != nil {
		t.Fatalf("marshal absent Optional: %v", err)
	}
	if got, want := string(absent), `{}`; got != want {
		t.Fatalf("marshal absent Optional = %s, want %s", got, want)
	}

	presentZero, err := json.Marshal(config{Retries: Optional[int]{Present: true, Value: 0}})
	if err != nil {
		t.Fatalf("marshal present-zero Optional: %v", err)
	}
	if got, want := string(presentZero), `{"retries":0}`; got != want {
		t.Fatalf("marshal present-zero Optional = %s, want %s", got, want)
	}

	var missing config
	if err := json.Unmarshal([]byte(`{}`), &missing); err != nil {
		t.Fatalf("unmarshal missing Optional: %v", err)
	}
	if missing.Retries.Present {
		t.Fatal("missing Optional is present")
	}

	var zero config
	if err := json.Unmarshal([]byte(`{"retries":0}`), &zero); err != nil {
		t.Fatalf("unmarshal present-zero Optional: %v", err)
	}
	if !zero.Retries.Present || zero.Retries.Value != 0 {
		t.Fatalf("present-zero Optional = %+v", zero.Retries)
	}
}

func TestOptionalPresentEmptyValue(t *testing.T) {
	type config struct {
		Name Optional[string] `json:"name,omitzero"`
	}
	data, err := json.Marshal(config{Name: Optional[string]{Present: true}})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `{"name":""}`; got != want {
		t.Fatalf("marshal present-empty Optional = %s, want %s", got, want)
	}
}

func TestOptionalRejectsNullAndPreservesReceiverOnFailure(t *testing.T) {
	original := Optional[int]{Present: true, Value: 7}
	for _, input := range []string{`null`, `"wrong"`} {
		got := original
		if err := json.Unmarshal([]byte(input), &got); err == nil {
			t.Fatalf("unmarshal %s unexpectedly succeeded", input)
		}
		if got != original {
			t.Fatalf("unmarshal %s mutated receiver: got %+v, want %+v", input, got, original)
		}
	}
}

func TestOptionalAbsentCannotMarshalStandalone(t *testing.T) {
	if _, err := json.Marshal(Optional[int]{}); err == nil {
		t.Fatal("marshal absent standalone Optional unexpectedly succeeded")
	}
}

func TestNullableJSONStates(t *testing.T) {
	type config struct {
		Timeout Nullable[int] `json:"timeout,omitzero"`
	}

	nullValue, err := json.Marshal(config{})
	if err != nil {
		t.Fatalf("marshal null Nullable: %v", err)
	}
	if got, want := string(nullValue), `{"timeout":null}`; got != want {
		t.Fatalf("marshal null Nullable = %s, want %s", got, want)
	}

	presentZero, err := json.Marshal(config{Timeout: Nullable[int]{Present: true, Value: 0}})
	if err != nil {
		t.Fatalf("marshal present-zero Nullable: %v", err)
	}
	if got, want := string(presentZero), `{"timeout":0}`; got != want {
		t.Fatalf("marshal present-zero Nullable = %s, want %s", got, want)
	}

	got := Nullable[int]{Present: true, Value: 7}
	if err := json.Unmarshal([]byte(`null`), &got); err != nil {
		t.Fatalf("unmarshal null Nullable: %v", err)
	}
	if got.Present || got.Value != 0 {
		t.Fatalf("null Nullable = %+v", got)
	}

	if err := json.Unmarshal([]byte(`0`), &got); err != nil {
		t.Fatalf("unmarshal present-zero Nullable: %v", err)
	}
	if !got.Present || got.Value != 0 {
		t.Fatalf("present-zero Nullable = %+v", got)
	}
}

func TestNullablePresentEmptyValue(t *testing.T) {
	value := Nullable[string]{Present: true}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `""`; got != want {
		t.Fatalf("marshal present-empty Nullable = %s, want %s", got, want)
	}
}

func TestNullablePreservesReceiverOnDecodeFailure(t *testing.T) {
	original := Nullable[int]{Present: true, Value: 7}
	got := original
	if err := json.Unmarshal([]byte(`"wrong"`), &got); err == nil {
		t.Fatal("invalid Nullable input unexpectedly succeeded")
	}
	if got != original {
		t.Fatalf("failed Nullable decode mutated receiver: got %+v, want %+v", got, original)
	}
}

func TestPresentNilCannotMarshalAsNull(t *testing.T) {
	var value *int
	if _, err := json.Marshal(Optional[*int]{Present: true, Value: value}); err == nil {
		t.Fatal("present nil Optional unexpectedly marshaled")
	}
	if _, err := json.Marshal(Nullable[*int]{Present: true, Value: value}); err == nil {
		t.Fatal("present nil Nullable unexpectedly marshaled")
	}
}

func TestPresentNonFiniteFloatCannotMarshal(t *testing.T) {
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := json.Marshal(Optional[float64]{Present: true, Value: value}); err == nil {
			t.Fatalf("present Optional %v unexpectedly marshaled", value)
		}
		if _, err := json.Marshal(Nullable[float64]{Present: true, Value: value}); err == nil {
			t.Fatalf("present Nullable %v unexpectedly marshaled", value)
		}
	}
}
