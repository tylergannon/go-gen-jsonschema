//go:build jsonschema
// +build jsonschema

package typescanner

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (TypeForSchemaMethod) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (*PointerTypeForSchemaMethod) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func TypeSchema() (json.RawMessage, error) {
	panic("not implemented")
}

func TypeSchema2() (json.RawMessage, error) {
	panic("not implemented")
}

var (
	_ = jsonschema.NewJSONSchemaMethod(TypeForSchemaMethod.Schema)
	_ = jsonschema.NewJSONSchemaMethod((*PointerTypeForSchemaMethod).Schema)
	_ = jsonschema.NewJSONSchemaBuilder[TypeForSchemaFunction](TypeSchema)
	_ = jsonschema.NewJSONSchemaBuilder[*PointerTypeForSchemaFunction](TypeSchema2)
	_ = jsonschema.NewInterfaceImpl[MarkerInterface](Type001{}, Type002{}, &Type003{}, (*Type004)(nil))
	_ = jsonschema.NewEnumType[NiceEnumType]()
)
