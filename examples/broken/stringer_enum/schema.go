//go:build jsonschema
// +build jsonschema

package stringer_enum

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Config) Schema() json.RawMessage { panic("not implemented") }

var _ = jsonschema.NewJSONSchemaMethod(
	Config.Schema,
	jsonschema.WithEnum(Config{}.LogLevel),
)

// PROBLEM: Even with a Stringer implementation,
// the generated schema uses integer values, not the string representations.
