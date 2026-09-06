//go:build jsonschema

package sealed_interface_slices

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Batch) Schema() json.RawMessage { panic("not implemented") }

// Event is sealed by its unexported isEvent method, so the slice element
// union (Created as a value variant, Deleted as a pointer variant) is
// inferred; nothing is declared at the field.
var _ = polytype.Declare(Batch.Schema)
