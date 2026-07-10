package defined_wrapper

import jsonschema "github.com/tylergannon/go-gen-jsonschema"

type MaybeInt jsonschema.Optional[int]

type Config struct {
	Value MaybeInt `json:"value,omitzero"`
}
