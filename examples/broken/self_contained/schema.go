//go:build jsonschema
// +build jsonschema

package self_contained

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Issue) Schema() json.RawMessage { panic("not implemented") }

// DESIRED: This should be sufficient - enums should be discovered from field options
var _ = jsonschema.NewJSONSchemaMethod(
	Issue.Schema,
	jsonschema.WithEnum(Issue{}.Priority),
	jsonschema.WithEnum(Issue{}.Severity),
)

// CURRENT REALITY: Must also have these redundant registrations
// Without these, the enums won't be properly generated
var (
	_ = jsonschema.NewEnumType[Priority]()
	_ = jsonschema.NewEnumType[Severity]()
)

// PROBLEM: The WithEnum options should make the global registrations unnecessary.
// The whole point of the Options pattern is to be self-contained!
