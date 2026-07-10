package nested_wrapper

import jsonschema "github.com/tylergannon/go-gen-jsonschema"

type Config struct {
	Value jsonschema.Optional[jsonschema.Optional[int]] `json:"value,omitzero"`
}
