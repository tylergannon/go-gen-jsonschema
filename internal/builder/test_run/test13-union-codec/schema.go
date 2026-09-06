//go:build jsonschema

package union_codec

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Envelope) Schema() json.RawMessage   { panic("not implemented") }
func (Envelope) ValidateJSON([]byte) error { panic("not implemented") }
func (Nested) Schema() json.RawMessage     { panic("not implemented") }

// Every Event field on Envelope and Nested shares the single inferred Event
// union; nothing about the union is declared here.
var _ = polytype.NewJSONSchemaMethod(
	Envelope.Schema,
	polytype.WithStringerEnum(Envelope{}.State),
)

var _ = polytype.NewJSONSchemaMethod(Nested.Schema)
