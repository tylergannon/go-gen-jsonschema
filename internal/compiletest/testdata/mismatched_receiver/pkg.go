package mismatched_receiver

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

type A struct{ X string }
type B struct{ Y int }

func (A) Schema() json.RawMessage    { panic("x") }
func (B) YSchema(int) json.Marshaler { panic("x") }

var _ = jsonschema.Declare(A.Schema).Method(A{}.X, B.YSchema)
