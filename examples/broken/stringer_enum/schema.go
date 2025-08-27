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
	// This SHOULD use the String() values ["DEBUG", "INFO", etc.]
	// but currently generates [0, 1, 2, 3, 4] instead
	jsonschema.WithEnum(Config{}.LogLevel),
	jsonschema.WithEnumMode(jsonschema.EnumStrings),
)

// PROBLEM: Even with WithEnumMode(EnumStrings) and a Stringer implementation,
// the generated schema uses integer values, not the string representations.
