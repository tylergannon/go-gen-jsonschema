//go:build jsonschema

package iota_global

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Task) Schema() json.RawMessage { panic("not implemented") }

var (
	_ = polytype.NewJSONSchemaMethod(Task.Schema)

	// THIS WILL PANIC: iota enums can't be registered globally
	// Error: panic: interface conversion: dst.Expr is *dst.Ident, not *dst.BasicLit
	// Location: internal/builder/gen_schema.go:454
	_ = polytype.NewEnumType[Priority]()
)
