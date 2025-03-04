//go:build jsonschema
// +build jsonschema

package typescanner

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"github.com/tylergannon/go-gen-jsonschema/internal/syntax/testfixtures/typescanner/scannersubpkg"
)

func (TypeForSchemaMethod) Schema() json.RawMessage {
	panic("not implemented")
}

func (*PointerTypeForSchemaMethod) Schema() json.RawMessage {
	panic("not implemented")
}

func TypeSchema() json.RawMessage {
	panic("not implemented")
}

func TypeSchema2() json.RawMessage {
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

var (
	_ = jsonschema.NewJSONSchemaBuilder[scannersubpkg.TypeForSchemaFunction](TypeSchema)
	_ = jsonschema.NewJSONSchemaBuilder[*scannersubpkg.PointerTypeForSchemaFunction](TypeSchema2)

	_ = jsonschema.NewInterfaceImpl[scannersubpkg.MarkerInterface](scannersubpkg.Type001{}, scannersubpkg.Type002{}, &scannersubpkg.Type003{}, (*scannersubpkg.Type004)(nil))

	_ = jsonschema.NewEnumType[scannersubpkg.NiceEnumType]()
)
