package stringer_enums

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestApplicationConfigUsesConstantNamesAndTaskStaysNumeric(t *testing.T) {
	config := ApplicationConfig{
		AppName:         "example",
		LogLevel:        LogWarning,
		DefaultPriority: PriorityHigh,
		MaxConnections:  4,
	}
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]json.RawMessage
	if err := json.Unmarshal(data, &wire); err != nil {
		t.Fatal(err)
	}
	assertEnumWire(t, wire["log_level"], `"LogWarning"`)
	assertEnumWire(t, wire["default_priority"], `"PriorityHigh"`)
	var decoded ApplicationConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, config) {
		t.Fatalf("decoded config = %#v, want %#v", decoded, config)
	}

	taskData, err := json.Marshal(Task{ID: "1", Name: "task", Priority: PriorityHigh, LogLevel: LogWarning})
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(taskData, &wire); err != nil {
		t.Fatal(err)
	}
	assertEnumWire(t, wire["priority"], "300")
	assertEnumWire(t, wire["log_level"], "2")
}

func assertEnumWire(t *testing.T, got json.RawMessage, want string) {
	t.Helper()
	if string(got) != want {
		t.Fatalf("wire value = %s, want %s", got, want)
	}
}
