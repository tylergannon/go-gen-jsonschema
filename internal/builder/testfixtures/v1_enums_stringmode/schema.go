//go:build jsonschema

package v1_enums_stringmode

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Paint) Schema() json.RawMessage { panic("not implemented") }

var _ = polytype.NewJSONSchemaMethod(
	Paint.Schema,
	polytype.WithStringerEnum(Paint{}.C),
	polytype.WithStringerEnum(Paint{}.Optional),
	polytype.WithStringerEnum(Paint{}.Nullable),
	polytype.WithEnum(Paint{}.Numeric),
	polytype.WithStringerEnum(Paint{}.Finish),
	polytype.WithStringerEnum(Paint{}.Remote),
)
