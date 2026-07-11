// Package ref_types is the acceptance example for AsRef(): a type registered
// with AsRef() is rendered as "$ref" into "$defs" wherever another
// registered schema references it, instead of being inlined.
package ref_types

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
