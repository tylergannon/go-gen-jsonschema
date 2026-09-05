//go:build jsonschema

package optionality

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Config) Schema() json.RawMessage        { panic("not implemented") }
func (NumericConfig) Schema() json.RawMessage { panic("not implemented") }

var _ = polytype.Declare(Config.Schema).
	Interface(Config{}.Pet, polytype.Discriminator("!kind"), polytype.Impl("Dog", Dog{}), polytype.Impl("Cat", Cat{}))

var _ = polytype.Declare(NumericConfig.Schema)
