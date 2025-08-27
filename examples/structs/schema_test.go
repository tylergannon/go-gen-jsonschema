package structs

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestPersonSchemaWithTime(t *testing.T) {
	// Get the generated schema
	person := Person{}
	schemaBytes := person.Schema()

	// Parse the schema
	var schema map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &schema); err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	// Check that birthDate is properly defined
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema missing properties field")
	}

	birthDate, ok := props["birthDate"].(map[string]interface{})
	if !ok {
		t.Fatal("Missing birthDate property")
	}

	// Verify it's a string type
	if birthDate["type"] != "string" {
		t.Errorf("birthDate type = %v, want string", birthDate["type"])
	}

	// Verify the description includes RFC3339 guidance
	desc, ok := birthDate["description"].(string)
	if !ok {
		t.Fatal("birthDate missing description")
	}

	if !strings.Contains(desc, "RFC3339") {
		t.Errorf("birthDate description doesn't mention RFC3339: %s", desc)
	}

	if !strings.Contains(desc, "2006-01-02T15:04:05Z07:00") {
		t.Errorf("birthDate description doesn't include example format: %s", desc)
	}
}

func TestPersonJSONMarshalUnmarshal(t *testing.T) {
	// Create a person with a specific time
	original := Person{
		ID:        "123",
		Name:      "John Doe",
		BirthDate: time.Date(1990, 5, 15, 14, 30, 0, 0, time.UTC),
		ContactInfo: ContactInfo{
			Email:           "john@example.com",
			Phone:           "555-1234",
			AlternateEmails: []string{"john.doe@example.com"},
		},
		Tags: []string{"developer", "golang"},
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal Person: %v", err)
	}

	// Verify the time is in RFC3339 format
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &jsonMap); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	birthDateStr, ok := jsonMap["birthDate"].(string)
	if !ok {
		t.Fatal("birthDate is not a string in JSON")
	}

	// Parse the time string to verify it's valid RFC3339
	parsedTime, err := time.Parse(time.RFC3339, birthDateStr)
	if err != nil {
		t.Errorf("birthDate is not valid RFC3339: %s, error: %v", birthDateStr, err)
	}

	// Unmarshal back to Person
	var unmarshaled Person
	if err := json.Unmarshal(jsonBytes, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal Person: %v", err)
	}

	// Verify the time matches (compare Unix timestamps to avoid timezone issues)
	if !original.BirthDate.Equal(unmarshaled.BirthDate) {
		t.Errorf("BirthDate mismatch: original=%v, unmarshaled=%v",
			original.BirthDate, unmarshaled.BirthDate)
	}

	// The unmarshaled time should match what we parsed
	if !parsedTime.Equal(unmarshaled.BirthDate) {
		t.Errorf("Parsed time doesn't match unmarshaled: parsed=%v, unmarshaled=%v",
			parsedTime, unmarshaled.BirthDate)
	}
}
