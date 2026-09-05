package nullable_provider

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

type Config struct {
	Value polytype.Nullable[string] `json:"value"`
}

func ValueSchema(polytype.Nullable[string]) json.Marshaler {
	return json.RawMessage(`{"type":"string"}`)
}
