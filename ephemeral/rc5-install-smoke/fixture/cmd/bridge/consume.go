package main

import (
	"encoding/json"
	"fmt"
	"os"

	transport "example.com/rc5-install-smoke"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func consume(input, output string) {
	data, err := os.ReadFile(input)
	check(err)
	check((transport.Shipment{}).ValidateJSON(data))

	var value transport.Shipment
	check(json.Unmarshal(data, &value))
	if value.ID != "shipment-42" || value.Status != transport.StatusDelivered || value.Quantity != 4 {
		fatalf("Go observed unexpected ordinary fields: %#v", value)
	}
	dispatched, ok := value.Event.(transport.Dispatched)
	if !ok || dispatched.Carrier != "Parcel Post" || dispatched.Tracking != "TRACK-42" {
		fatalf("Go observed unexpected union value: %#v", value.Event)
	}
	if !value.Note.Present || value.Note.Value != "" {
		fatalf("Go did not preserve the present empty Optional value: %#v", value.Note)
	}
	if value.ETA.Present {
		fatalf("Go did not decode the Nullable null state: %#v", value.ETA)
	}

	value.Quantity++
	value.Note = jsonschema.Optional[string]{Present: true, Value: "accepted-by-go"}
	value.ETA = jsonschema.Nullable[string]{Present: true, Value: "2026-09-05T16:30:00Z"}
	encoded, err := json.MarshalIndent(value, "", "  ")
	check(err)
	check((transport.Shipment{}).ValidateJSON(encoded))
	check(os.WriteFile(output, append(encoded, '\n'), 0o644))
	fmt.Println("go_consume_ok status=StatusDelivered kind=dispatched optional_empty_observed=true nullable_null_observed=true reencoded_quantity=5")
}
