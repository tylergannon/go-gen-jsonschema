package nullable_slice

import jsonschema "github.com/tylergannon/go-gen-jsonschema"

type Config struct {
	Value jsonschema.Nullable[[]int] `json:"value"`
}
