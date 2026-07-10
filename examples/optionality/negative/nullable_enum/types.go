package nullable_enum

import jsonschema "github.com/tylergannon/go-gen-jsonschema"

type Color string

const (
	Red  Color = "red"
	Blue Color = "blue"
)

type Config struct {
	Shade jsonschema.Nullable[Color] `json:"shade"`
}
