//go:build jsonschema

package test_options

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

// Schema method stubs for iota enum types
func (Product) Schema() json.RawMessage       { panic("not implemented") }
func (Configuration) Schema() json.RawMessage { panic("not implemented") }

// Color, Size, and LogLevel declare `func (T) enum()` in enum_iota_types.go
// and are emitted as integer enums of their iota values. A String() method
// on a marked type is ignored; use .StringerEnum on a field to emit constant
// names instead.
var _ = polytype.NewJSONSchemaMethod(Product.Schema)
var _ = polytype.NewJSONSchemaMethod(Configuration.Schema)
