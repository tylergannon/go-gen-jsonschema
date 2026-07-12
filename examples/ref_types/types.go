// Package ref_types is the acceptance example for AsRef(): a type registered
// with AsRef() is rendered as "$ref" into "$defs" wherever another
// registered schema references it, instead of being inlined.
package ref_types

import jsonschema "github.com/tylergannon/go-gen-jsonschema"

//go:generate go run ../../gen-jsonschema/ --pretty --validate

// Shared is registered with AsRef(). Wherever another registered schema
// references it, it appears as a "$ref" into that schema's "$defs" instead
// of being inlined.
type Shared struct {
	Name string `json:"name"`
}

// Container references Shared twice: once directly, once inside a slice.
// Both references collapse to the same "$defs" entry.
type Container struct {
	Primary Shared   `json:"primary"`
	Others  []Shared `json:"others"`
}

// Mode is a registered string enum used to prove nullable enum rendering.
type Mode string

const (
	ModeFast Mode = "fast"
	ModeSafe Mode = "safe"
)

// NullableConfig exercises the two nullable shapes that retain reusable
// contracts without widening Nullable support to arbitrary schema nodes.
type NullableConfig struct {
	Mode   jsonschema.Nullable[Mode]   `json:"mode"`
	Shared jsonschema.Nullable[Shared] `json:"shared"`
}
