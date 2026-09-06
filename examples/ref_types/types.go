// Package ref_types is the acceptance example for Ref(): a type registered
// with Ref() is rendered as "$ref" into "$defs" wherever another
// registered schema references it, instead of being inlined.
package ref_types

import "github.com/tylergannon/polytype"

//go:generate go run ../../polytype/ --pretty --validate

// Shared is registered with Ref(). Wherever another registered schema
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
	Mode   polytype.Nullable[Mode]   `json:"mode"`
	Shared polytype.Nullable[Shared] `json:"shared"`
}

// Mode declares itself as an enum; the generator emits its typed constants.
func (Mode) enum() {}
