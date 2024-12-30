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
	file      *dst.File
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

func NewFuncEntry(decl *dst.FuncDecl, pkg *decorator.Package, file *dst.File, importMap ImportMap) *FuncEntry {
	typeID := NewTypeID(pkg.PkgPath, funcNameFromDst(decl))
	return &FuncEntry{ImportMap: importMap, FuncDecl: decl, typeID: typeID, file: file}
}

func (fe *FuncEntry) isCandidateAltConverter() bool {
	// Check function arguments
	params := fe.FuncDecl.Type.Params
	results := fe.FuncDecl.Type.Results

	if fe.FuncDecl.Recv == nil {
		if len(params.List) != 1 {
			return false
		}
	} else if _, ok := fe.FuncDecl.Recv.List[0].Type.(*dst.StarExpr); ok {
		return false
	} else if len(params.List) != 0 {
		return false
	}

	if results == nil || len(results.List) != 2 {
		return false
	}
	var ident, ok = results.List[1].Type.(*dst.Ident)
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

func (fe *FuncEntry) IsUnmarshalJSON() (typeName, pkgPath string, yesItIs bool) {
	// 1. The function must be named "UnmarshalJSON".
	if fe.Func.Name() != "UnmarshalJSON" {
		return "", "", false
	}

	sig, ok := fe.Func.Type().(*types.Signature)
	if !ok {
		return "", "", false
	}

	// 2. Exactly one parameter of type []byte.
	params := sig.Params()
	if params.Len() != 1 {
		return "", "", false
	}

	paramType := params.At(0).Type()
	sliceType, ok := paramType.(*types.Slice)
	if !ok {
		return "", "", false
	}
	elemType := sliceType.Elem()
	// `[]byte` is actually `[]uint8` in go/types
	if basic, ok := elemType.(*types.Basic); !ok || basic.Kind() != types.Byte {
		return "", "", false
	}

	// 3. Exactly one result, of type error.
	results := sig.Results()
	if results.Len() != 1 {
		return "", "", false
	}
	resultType := results.At(0).Type()
	// The standard "error" type is checked by string match or using IsError check:
	// if resultType.String() != "error" { ... }
	// or we can do a more robust check as below:
	if !types.Implements(types.NewPointer(resultType), types.Universe.Lookup("error").Type().Underlying().(*types.Interface)) &&
		!types.Implements(resultType, types.Universe.Lookup("error").Type().Underlying().(*types.Interface)) {
		return "", "", false
	}

	// 4. The receiver must be a method, so extract the receiver’s type.
	recv := sig.Recv()
	if recv == nil {
		return "", "", false
	}
	recvType := recv.Type()

	// If the receiver is a pointer type, unwrap it to get the Named type.
	if ptr, ok := recvType.(*types.Pointer); ok {
		recvType = ptr.Elem()
	}

	// Now recvType should (typically) be a *types.Named
	named, ok := recvType.(*types.Named)
	if !ok {
		return "", "", false
	}

	// Extract type name and package path
	typeName = named.Obj().Name()
	if named.Obj().Pkg() != nil {
		pkgPath = named.Obj().Pkg().Path()
	}

	return typeName, pkgPath, true
}
