package nullable_ref

import jsonschema "github.com/tylergannon/go-gen-jsonschema"

type Config struct {
	Value jsonschema.Nullable[string] `json:"value" jsonschema:"ref=definitions/Value"`
}
