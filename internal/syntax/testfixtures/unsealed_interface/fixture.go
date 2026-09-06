//go:build jsonschema

// Package unsealedinterface is a standalone negative fixture: a generated
// schema reaches an interface field whose interface declares no unexported
// method, so its membership cannot be inferred. That must be a hard,
// source-positioned diagnostic naming the interface and the field rather
// than a silent fallback. It lives in its own directory because loading a
// package eagerly scans every type in it.
package unsealedinterface

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

// Shape is not sealed: Area is exported, so anything anywhere could
// implement it.
type Shape interface {
	Area() float64
}

type Circle struct {
	Radius float64 `json:"radius"`
}

func (c Circle) Area() float64 { return 3 * c.Radius * c.Radius }

type Drawing struct {
	Shape Shape `json:"shape"`
}

func (Drawing) Schema() json.RawMessage { panic("not implemented") }

var _ = polytype.Declare(Drawing.Schema)
