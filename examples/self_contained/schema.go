//go:build jsonschema

package self_contained

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (Issue) Schema() json.RawMessage { panic("not implemented") }

// DESIRED: This should be sufficient - enums should be discovered from field options
var _ = polytype.NewJSONSchemaMethod(
	Issue.Schema,
	polytype.WithEnum(Issue{}.Priority),
	polytype.WithEnum(Issue{}.Severity),
)

// CURRENT REALITY: Must also have these redundant registrations
// Without these, the enums won't be properly generated
// var (
// 	_ = polytype.NewEnumType[Priority]()
// 	_ = polytype.NewEnumType[Severity]()
// )

// PROBLEM: The WithEnum options should make the global registrations unnecessary.
// The whole point of the Options pattern is to be self-contained!
