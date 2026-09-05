//go:build jsonschema

package typescanner

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
	"github.com/tylergannon/polytype/internal/syntax/testfixtures/typescanner/scannersubpkg"
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
	_ = polytype.NewJSONSchemaMethod(TypeForSchemaMethod.Schema)
	_ = polytype.NewJSONSchemaMethod((*PointerTypeForSchemaMethod).Schema)
	_ = polytype.NewJSONSchemaBuilder[TypeForSchemaFunction](TypeSchema)
	_ = polytype.NewJSONSchemaBuilder[*PointerTypeForSchemaFunction](TypeSchema2)
	_ = polytype.NewInterfaceImpl[MarkerInterface](Type001{}, Type002{}, &Type003{}, (*Type004)(nil))
	_ = polytype.NewEnumType[NiceEnumType]()
)

var (
	_ = polytype.NewJSONSchemaBuilder[scannersubpkg.TypeForSchemaFunction](TypeSchema)
	_ = polytype.NewJSONSchemaBuilder[*scannersubpkg.PointerTypeForSchemaFunction](TypeSchema2)

	_ = polytype.NewInterfaceImpl[scannersubpkg.MarkerInterface](scannersubpkg.Type001{}, scannersubpkg.Type002{}, &scannersubpkg.Type003{}, (*scannersubpkg.Type004)(nil))

	_ = polytype.NewEnumType[scannersubpkg.NiceEnumType]()
)
