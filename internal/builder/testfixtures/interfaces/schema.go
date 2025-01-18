//go:build jsonschema
// +build jsonschema

package interfaces

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (FancyStruct) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

var (
	_ = jsonschema.NewJSONSchemaMethod(FancyStruct.Schema)
	_ = jsonschema.NewInterfaceImpl[TestInterface](TestInterface1{}, TestInterface2{})
)
