//go:build jsonschema

package template_rendering

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (WorkItem) Schema() json.RawMessage { panic("not implemented") }

// Status declares `func (Status) enum()` in types.go, so WorkItem.Status is
// emitted as an enum with no field-level registration.
var _ = polytype.Declare(WorkItem.Schema)
