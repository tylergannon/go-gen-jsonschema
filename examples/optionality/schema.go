//go:build jsonschema

package optionality

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Config) Schema() json.RawMessage        { panic("not implemented") }
func (NumericConfig) Schema() json.RawMessage { panic("not implemented") }

var _ = jsonschema.NewJSONSchemaMethod(
	Config.Schema,
	jsonschema.WithInterface(Config{}.Pet),
	jsonschema.WithInterfaceImpls(Config{}.Pet, Dog{}, Cat{}),
	jsonschema.WithDiscriminator(Config{}.Pet, "!kind"),
)

var _ = jsonschema.NewJSONSchemaMethod(NumericConfig.Schema)
