package optional_without_omitzero

import jsonschema "github.com/tylergannon/go-gen-jsonschema"

type Config struct {
	Value jsonschema.Optional[int] `json:"value"`
}
