//go:build jsonschema

package optionality

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Config) Schema() json.RawMessage        { panic("not implemented") }
func (NumericConfig) Schema() json.RawMessage { panic("not implemented") }

// Pet is sealed by its unexported pet method; Dog and Cat are inferred.
var _ = polytype.Declare(Config.Schema)

var _ = polytype.SealedUnion[Pet]("!kind")

var _ = polytype.Declare(NumericConfig.Schema)
