//go:build jsonschema

package consumer

import (
	"encoding/json"
	"github.com/tylergannon/polytype"
)

func (Envelope) Schema() json.RawMessage    { panic("not implemented") }
func (Detail) Schema() json.RawMessage      { panic("not implemented") }
func (Composition) Schema() json.RawMessage { panic("not implemented") }

var _ = polytype.NewJSONSchemaMethod(Detail.Schema, polytype.AsRef())
var _ = polytype.NewJSONSchemaMethod(Composition.Schema)
var _ = polytype.NewJSONSchemaMethod(Envelope.Schema,
	polytype.WithStringerEnum(Envelope{}.PriorityName),
)
