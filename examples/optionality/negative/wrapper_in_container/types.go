package wrapper_in_container

import jsonschema "github.com/tylergannon/go-gen-jsonschema"

type Config struct {
	Values []jsonschema.Optional[int] `json:"values"`
}
