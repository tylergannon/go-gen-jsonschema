package main

import (
	"encoding/json"
	"fmt"
	"os"

	transport "example.com/adoption-transport-proof"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func main() {
	switch {
	case len(os.Args) == 3 && os.Args[1] == "emit":
		emit(os.Args[2])
	case len(os.Args) == 4 && os.Args[1] == "consume":
		consume(os.Args[2], os.Args[3])
	default:
		fatalf("usage: bridge emit OUTPUT | bridge consume INPUT OUTPUT")
	}
}

func emit(output string) {
	value := transport.Shipment{
		ID:       "shipment-42",
		Status:   transport.StatusReady,
		Event:    transport.Created{Actor: "go-service", At: "2026-09-04T18:30:00Z"},
		Note:     jsonschema.Optional[string]{Present: true, Value: "packed"},
		ETA:      jsonschema.Nullable[string]{Present: true, Value: "2026-09-05T14:00:00Z"},
		Quantity: 3,
	}
	data, err := json.MarshalIndent(value, "", "  ")
	check(err)
	check((transport.Shipment{}).ValidateJSON(data))
	check(os.WriteFile(output, append(data, '\n'), 0o644))
	fmt.Println("go_emit_ok status=StatusReady kind=created note_present=true eta_present=true quantity=3")
}

func check(err error) {
	if err != nil {
		fatalf("%v", err)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
