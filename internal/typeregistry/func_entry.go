package typeregistry

import (
	"fmt"
	"github.com/dave/dst"
	"go/types"
)

type FuncEntry struct {
	*types.Func
	*dst.FuncDecl
	typeID    TypeID
	ImportMap ImportMap
}

func (fe *FuncEntry) TakesPtr() bool {
	if fe.FuncDecl.Recv != nil && len(fe.FuncDecl.Recv.List) > 0 {
		return false
	}
	paramExpr := fe.FuncDecl.Type.Params.List[0].Type
	_, ok := paramExpr.(*dst.StarExpr)
	return ok
}

func (fe *FuncEntry) ReceiverOrArgType() (pkgPath, typeName string) {
	// Check for a receiver first
	if fe.FuncDecl.Recv != nil && len(fe.FuncDecl.Recv.List) > 0 {
		recv := fe.FuncDecl.Recv.List[0].Type

		ident := recv.(*dst.Ident)
		return fe.ImportMap[""], ident.Name
	}

	var (
		importAlias = ""
	)

	paramExpr := fe.FuncDecl.Type.Params.List[0].Type
	if starExpr, ok := paramExpr.(*dst.StarExpr); ok {
		paramExpr = starExpr.X
	}
	switch t := paramExpr.(type) {
	case *dst.Ident:
		typeName = t.Name
		if t.Path == "" {
			return fe.ImportMap[""], typeName
		}
		return t.Path, typeName
	case *dst.SelectorExpr:
		x := t.X.(*dst.Ident)
		fmt.Printf("The selector is X.Sel: %s.%s", x.Name, t.Sel.Name)
		importAlias = x.Name
		typeName = t.Sel.Name
	}
	return fe.ImportMap[importAlias], typeName
}
