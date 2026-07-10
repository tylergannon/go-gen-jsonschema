package nullable_provider

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

type Config struct {
	Value jsonschema.Nullable[string] `json:"value"`
}

func ValueSchema(jsonschema.Nullable[string]) json.Marshaler {
	return json.RawMessage(`{"type":"string"}`)
}
