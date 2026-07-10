package embedded_wrapper

import jsonschema "github.com/tylergannon/go-gen-jsonschema"

type Config struct {
	jsonschema.Optional[int]
}
