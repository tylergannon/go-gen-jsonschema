//go:build jsonschema

package optionality

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Config) Schema() json.RawMessage        { panic("not implemented") }
func (NumericConfig) Schema() json.RawMessage { panic("not implemented") }

var _ = jsonschema.Declare(Config.Schema).
	Interface(Config{}.Pet, jsonschema.Discriminator("!kind"), jsonschema.Impl("Dog", Dog{}), jsonschema.Impl("Cat", Cat{}))

var _ = jsonschema.Declare(NumericConfig.Schema)
