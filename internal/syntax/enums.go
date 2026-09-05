package syntax

import (
	"fmt"
	"go/token"
	"go/types"

	"github.com/dave/dst"
)

// ResolveEnum returns the exact package-level constants whose Go type is the
// provided named enum type, retaining declaration order and constant names.
// Untyped constants are excluded; constants typed through conversions and
// evaluated expressions are included by go/types identity.
func ResolveEnum(typeSpec TypeSpec) (*EnumSet, error) {
	object, ok := typeSpec.Pkg().Types.Scope().Lookup(typeSpec.Name()).(*types.TypeName)
	if !ok {
		return nil, fmt.Errorf("enum type %s was not resolved by go/types at %s", typeSpec.Name(), typeSpec.Position())
	}
	targetType := object.Type()
	if object.IsAlias() {
		targetType = types.Unalias(targetType)
		if _, ok := targetType.(*types.Named); !ok {
			return nil, fmt.Errorf("enum type %s at %s aliases a predeclared type; exact constant membership is not distinguishable from unrelated constants", typeSpec.Name(), typeSpec.Position())
		}
	}

	result := &EnumSet{TypeSpec: typeSpec}
	for _, file := range typeSpec.Pkg().Syntax {
		for _, declaration := range file.Decls {
			constantDeclaration, ok := declaration.(*dst.GenDecl)
			if !ok || constantDeclaration.Tok != token.CONST {
				continue
			}
			for _, specification := range constantDeclaration.Specs {
				valueSpec := NewValueSpec(constantDeclaration, specification.(*dst.ValueSpec), typeSpec.Pkg(), file)
				for _, ident := range valueSpec.Value().Names {
					constantObject, ok := typeSpec.Pkg().Types.Scope().Lookup(ident.Name).(*types.Const)
					if !ok || !types.Identical(types.Unalias(constantObject.Type()), targetType) {
						continue
					}
					result.Values = append(result.Values, EnumValue{
						Name:        ident.Name,
						Value:       constantObject.Val(),
						Description: valueSpec.Comments(),
						Source:      valueSpec.Position(),
					})
				}
			}
		}
	}
	return result, nil
}
