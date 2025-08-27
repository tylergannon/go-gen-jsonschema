//go:build jsonschema
// +build jsonschema

package custom_names

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Theme) Schema() json.RawMessage { panic("not implemented") }

var _ = jsonschema.NewJSONSchemaMethod(
	Theme.Schema,
	jsonschema.WithEnum(Theme{}.PrimaryColor),
	jsonschema.WithEnumMode(jsonschema.EnumStrings),
	// These custom names should be used in the schema
	jsonschema.WithEnumName(ColorRed, "red"),
	jsonschema.WithEnumName(ColorGreen, "green"),
	jsonschema.WithEnumName(ColorBlue, "blue"),
	jsonschema.WithEnumName(ColorYellow, "yellow"),

	// Same for secondary color
	jsonschema.WithEnum(Theme{}.SecondaryColor),
)

// PROBLEM: WithEnumName has no effect. The custom names are not applied
// to the generated schema. Should generate ["red", "green", "blue", "yellow"]
// but likely generates [0, 1, 2, 3] or doesn't work at all.
