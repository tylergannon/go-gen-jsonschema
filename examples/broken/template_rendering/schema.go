//go:build jsonschema
// +build jsonschema

package template_rendering

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (WorkItem) Schema() json.RawMessage { panic("not implemented") }

var (
	// Use WithEnum option to configure field-level enum
	_ = jsonschema.NewJSONSchemaMethod(
		WorkItem.Schema,
		jsonschema.WithEnum(WorkItem{}.Status),
	)

	// Also need this for now (shouldn't be required!)
	_ = jsonschema.NewEnumType[Status]()
)

// PROBLEM: This generates jsonschema/WorkItem.json.tmpl with {{.status}}
// But Schema() method just returns the raw template, not rendered JSON!
// The template is never rendered with actual enum values.
