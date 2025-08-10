//go:build jsonschema

package entrypoints

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (MethodType) Schema() json.RawMessage { panic("not implemented") }

func FuncTypeSchema(FuncType) json.RawMessage { panic("not implemented") }

func BuilderTypeSchema() json.RawMessage { panic("not implemented") }

var (
	_ = jsonschema.NewJSONSchemaMethod(MethodType.Schema)
	_ = jsonschema.NewJSONSchemaFunc[FuncType](FuncTypeSchema)
	_ = jsonschema.NewJSONSchemaBuilder[BuilderType](BuilderTypeSchema)
)
