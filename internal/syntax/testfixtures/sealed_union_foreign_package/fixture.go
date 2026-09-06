//go:build jsonschema

// Package sealedunionforeignpackage is a standalone negative fixture: it
// declares SealedUnion for an interface that another package declares. That
// must be a hard, source-positioned diagnostic naming the interface and the
// package that owns it. It lives in its own directory because loading a
// package eagerly parses every marker call in it.
package sealedunionforeignpackage

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
	"github.com/tylergannon/polytype/internal/syntax/testfixtures/sealed_union_foreign_package/animals"
)

type Zoo struct {
	Resident animals.Animal `json:"resident"`
}

func (Zoo) Schema() json.RawMessage { panic("not implemented") }

var _ = polytype.Declare(Zoo.Schema)

var _ = polytype.SealedUnion[animals.Animal]("kind")
