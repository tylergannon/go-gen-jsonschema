package typeregistry

import (
	"errors"
	"fmt"
	"github.com/dave/dst"
	"go/token"
)

var (
	ErrInvalidUnionTypeArg = errors.New("invalid union type arg")
)

// TypeAlternative is overloaded.
// It covers two cases with basically different fields.
// The first case is when a conversion function is used to change
type TypeAlternative struct {
	Alias          string    // unchanged
	ConversionFunc string    // unchanged
	ImportMap      ImportMap // unchanged

	// new fields for interface implementations
	InterfaceImpl bool // unchanged
	PackageName   string
	PkgPath       string
	TypeName      string
	IsPointer     bool
}

func (r *Registry) interpretUnionTypeAltArg(expr dst.Expr, importMap ImportMap) (alt TypeAlternative, err error) {
	callExpr, ok := expr.(*dst.CallExpr)
	if !ok {
		return alt, ErrInvalidUnionTypeArg
	}
	switch fun := callExpr.Fun.(type) {
	case *dst.Ident:
		if fun.Name != TypeAltFunc || fun.Path != JSONSchemaPackage {
			return alt, ErrInvalidUnionTypeArg
		}
	default:
		return alt, ErrInvalidUnionTypeArg
	}
	alt.Alias = callExpr.Args[0].(*dst.BasicLit).Value
	alt.Alias = alt.Alias[1 : len(alt.Alias)-1]
	alt.ImportMap = importMap

	switch typeArg := callExpr.Args[1].(type) {
	case *dst.SelectorExpr:
		// This is the case of a struct method whose receiver is the
		// alternate type.
		if ident, ok := typeArg.X.(*dst.Ident); ok {
			alt.ConversionFunc = fmt.Sprintf("%s.%s", ident.Name, typeArg.Sel.Name)
		} else {
			return alt, ErrInvalidUnionTypeArg
		}
	case *dst.Ident:
		alt.ConversionFunc = typeArg.Name
	default:
		return alt, ErrInvalidUnionTypeArg
	}

	return alt, nil
}

// interpretImplementationsArg interprets a single argument to SetImplementations,
// such as MyStruct{}, &MyStruct{}, (*MyStruct)(nil), etc.
//
// If the expression's type cannot be determined as a struct or pointer to a struct,
// an error is returned.
func (r *Registry) interpretImplementationsArg(expr dst.Expr, importMap ImportMap) (TypeAlternative, error) {
	var alt TypeAlternative
	alt.ImportMap = importMap
	alt.InterfaceImpl = true // this marks that it came from SetImplementations

	// We'll capture the final parse results
	var (
		pkgName   string
		typeName  string
		isPointer bool
	)

	switch e := expr.(type) {
	case *dst.CompositeLit:
		// e.Type is the type expression: e.g. "MyStruct" or "pkg.MyStruct"
		pkgName, typeName, isPointer, _ = parseTypeExpr(e.Type)
		if typeName == "" {
			return alt, errors.New("could not parse composite literal type")
		}

	case *dst.UnaryExpr:
		// Example: &MyStruct{}
		if e.Op != token.AND {
			return alt, fmt.Errorf("unhandled unary expression: %#v", e)
		}
		compLit, ok := e.X.(*dst.CompositeLit)
		if !ok {
			return alt, errors.New("unhandled expression form after '&'")
		}
		// e.X.Type is the type expression
		pkgName, typeName, isPointer, _ = parseTypeExpr(compLit.Type)
		if typeName == "" {
			return alt, errors.New("could not parse pointer composite literal type")
		}
		isPointer = true

	case *dst.CallExpr:
		// Example: (*MyStruct)(nil)
		// The Fun part is typically a ParenExpr(StarExpr(Ident))
		parenExpr, ok := e.Fun.(*dst.ParenExpr)
		if !ok {
			return alt, errors.New("call expression was not a cast: missing paren expr")
		}
		starExpr, ok := parenExpr.X.(*dst.StarExpr)
		if !ok {
			return alt, errors.New("call expression was not a pointer cast: missing star expr")
		}

		pkgName, typeName, _, _ = parseTypeExpr(starExpr.X)
		if typeName == "" {
			return alt, errors.New("could not parse type in pointer cast")
		}
		isPointer = true

	default:
		return alt, fmt.Errorf("unhandled expression kind for interface impl: %T", expr)
	}

	if pkgName != "" {
		var found bool
		if alt.PkgPath, found = importMap[pkgName]; !found {
			return alt, fmt.Errorf("unable to resolve path for package name %s for typeName %s", pkgName, typeName)
		}
	}

	alt.PackageName = pkgName
	alt.TypeName = typeName
	alt.IsPointer = isPointer
	return alt, nil
}

// parseTypeExpr is a small helper that drills into a dst.Expr to figure out
// (packageName, typeName, isPointer).  For example, it can parse:
//   - Ident("MyStruct") => ("", "MyStruct", false)
//   - SelectorExpr(X=Ident("impls"), Sel=Ident("MyStruct")) => ("impls", "MyStruct", false)
//   - StarExpr(X=Ident("MyStruct")) => ("", "MyStruct", true) // though we typically parse this outside
func parseTypeExpr(expr dst.Expr) (pkgName, typeName string, isPointer bool, err error) {
	switch typ := expr.(type) {
	case *dst.Ident:
		// e.g. "MyStruct"
		pkgName = ""
		typeName = typ.Name
		isPointer = false
	case *dst.SelectorExpr:
		// e.g. pkg.MyStruct
		pkgIdent, ok := typ.X.(*dst.Ident)
		if !ok {
			return "", "", false, errors.New("selector X was not an Ident")
		}
		pkgName = pkgIdent.Name
		typeName = typ.Sel.Name
		isPointer = false
	case *dst.StarExpr:
		// e.g. *MyStruct (rare inside parseTypeExpr, but let's handle it)
		pkgName, typeName, _, err = parseTypeExpr(typ.X)
		if err != nil {
			return "", "", false, err
		}
		isPointer = true
	default:
		return "", "", false, errors.New("unhandled type expression form")
	}
	return pkgName, typeName, isPointer, nil
}
