package mismatched_accessor

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

type Example struct{ A string }
type Other struct{ B string }

func (Example) Schema() json.RawMessage    { panic("x") }
func (Other) OtherASchema() json.Marshaler { panic("x") }

var _ = jsonschema.Declare(Example.Schema).Accessor(Example{}.A, Other.OtherASchema)
