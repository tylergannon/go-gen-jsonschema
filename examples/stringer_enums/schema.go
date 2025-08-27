//go:build jsonschema

package stringer_enums

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

// ApplicationConfig schema with WithStringerEnum (WITHOUT NewEnumType - testing auto-discovery!)
func (ApplicationConfig) Schema() json.RawMessage { panic("not implemented") }

var _ = jsonschema.NewJSONSchemaMethod(
	ApplicationConfig.Schema,
	jsonschema.WithStringerEnum(ApplicationConfig{}.LogLevel),
	jsonschema.WithStringerEnum(ApplicationConfig{}.DefaultPriority),
)

// Task schema with regular WithEnum (also WITHOUT NewEnumType!)
func (Task) Schema() json.RawMessage { panic("not implemented") }

var _ = jsonschema.NewJSONSchemaMethod(
	Task.Schema,
	jsonschema.WithEnum(Task{}.Priority), // This will use integer values
	jsonschema.WithEnum(Task{}.LogLevel), // This will use integer values
)
