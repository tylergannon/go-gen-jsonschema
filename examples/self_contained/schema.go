//go:build jsonschema

package self_contained

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Issue) Schema() json.RawMessage { panic("not implemented") }

// Priority and Severity declare `func (T) enum()` in types.go, so the root
// declaration alone is sufficient: no field-level or package-level enum
// registration is needed.
var _ = polytype.Declare(Issue.Schema)
