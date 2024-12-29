package typeregistry

import (
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"go/types"
)

// FuncEntry holds function declaration information for functions specifically
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

func NewFuncEntry(decl *dst.FuncDecl, pkg *decorator.Package, importMap ImportMap) *FuncEntry {
	typeID := NewTypeID(pkg.PkgPath, funcNameFromDst(decl))
	return &FuncEntry{ImportMap: importMap, FuncDecl: decl, typeID: typeID}
}

func (fe *FuncEntry) isCandidateAltConverter() bool {
	// Check function arguments
	params := fe.FuncDecl.Type.Params
	results := fe.FuncDecl.Type.Results

	fmt.Println(fe.typeID)
	if fe.FuncDecl.Recv == nil {
		if len(params.List) != 1 {
			fmt.Println("Not one")
			return false
		}
	} else if _, ok := fe.FuncDecl.Recv.List[0].Type.(*dst.StarExpr); ok {
		fmt.Println("Not two")
		return false
	} else if len(params.List) != 0 {
		return false
	}

	if results == nil || len(results.List) != 2 {
		fmt.Println("Not three")
		return false
	}
	var ident, ok = results.List[1].Type.(*dst.Ident)
	fmt.Printf("The ident is %v\n", ok && ident.Name == "error")
	return ok && ident.Name == "error"
}

// funcNameFromDst returns the name of the function if the function takes no
// receiver.  If it takes a receiver, the function will be namespaced with the
// type name of the receiver.
func funcNameFromDst(funcDecl *dst.FuncDecl) string {
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		if ident, ok := funcDecl.Recv.List[0].Type.(*dst.Ident); ok {
			return ident.Name + "." + funcDecl.Name.Name
		}
	}
	return funcDecl.Name.Name
}
