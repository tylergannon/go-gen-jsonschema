package scanner

import (
	"errors"
	"fmt"
	"github.com/tylergannon/go-gen-jsonschema/internal/importmap"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
	"strings"
)

type (
	// Indirection labels a TypeID to tell whether the indicated type is a
	// concrete instance of a named type, a pointer to it, etc.
	Indirection int

	// MarkerFunction is an enum that labels a marker function call to denote
	// which function was called.
	MarkerFunction string
)

const (
	MarkerFuncNewJSONSchemaBuilder MarkerFunction = "NewJSONSchemaBuilder" // NewJSONSchemaBuilder
	MarkerFuncNewJSONSchemaMethod  MarkerFunction = "NewJSONSchemaMethod"  // NewJSONSchemaMethod
	MarkerFuncNewInterfaceImpl     MarkerFunction = "NewInterfaceImpl"     // NewInterfaceImpl
	MarkerFuncNewEnumType          MarkerFunction = "NewEnumType"          // NewEnumType
)

const (
	// NormalConcrete The type named on the TypeID is just the basic or named type.
	NormalConcrete Indirection = iota
	// Pointer It's a pointer to the basic or named type.
	Pointer
	// SliceOfConcrete means the type is a slice of some named type or basic type.
	// NOTE: hypothetical.  not currently supported.
	SliceOfConcrete
	// SliceOfPointer means the type is a slice of pointer to a named or basic type.
	// NOTE: hypothetical.  not currently supported.
	SliceOfPointer
)

// TypeID is our structured representation of a type. It can represent named types,
// pointers, slices, arrays, and generics—plus marks invalid or disallowed types.
type (
	TypeID struct {
		// May be empty string if there is no package info available, meaning the type
		// is defined in the current package from where it's referenced.
		PkgPath  string
		TypeName string
		// MUST be set to true if the PkgPath is empty.
		DeclaredLocally bool
		Indirection     Indirection
	}

	// MarkerFunctionCall denotes a call to one of the marker functions, found
	// in the scanned source code.
	MarkerFunctionCall struct {
		Pkg      *packages.Package
		Function MarkerFunction
		// Our function calls need either zero or one type argument.
		// If present, denote the type argument here.
		TypeArgument *TypeID
		Arguments    []ast.Expr
		File         *ast.File
		Position     token.Position
	}
)

func (m MarkerFunctionCall) String() string {
	args := make([]string, len(m.Arguments))
	for i, arg := range m.Arguments {
		args[i] = fmt.Sprint(arg)
	}
	return fmt.Sprintf("%s %s Args{%s}", m.Function, m.TypeArgument, strings.Join(args, ","))
}

func (t TypeID) String() string {
	var (
		ptr     string
		pkgPath = t.PkgPath
	)
	if t.Indirection == Pointer {
		ptr = "*"
	}
	if pkgPath == "" {
		pkgPath = "<local>"
	}
	return fmt.Sprintf("%s%s.%s", ptr, pkgPath, t.TypeName)
}

func ParseValueExprForMarkerFunctionCall(e *ast.ValueSpec, file *ast.File, pkg *packages.Package) []MarkerFunctionCall {
	var results []MarkerFunctionCall
	for _, arg := range e.Values {
		ce, ok := arg.(*ast.CallExpr)
		if !ok {
			continue
		}
		id := parseFuncFromExpr(ce.Fun, file.Imports)
		if id.PkgPath != importmap.SchemaPackagePath {
			fmt.Println("Not path", id)
			continue
		}
		switch MarkerFunction(id.TypeName) {
		case MarkerFuncNewJSONSchemaBuilder, MarkerFuncNewJSONSchemaMethod, MarkerFuncNewInterfaceImpl, MarkerFuncNewEnumType:
		default:
			fmt.Println("Unsupported MarkerFunction", id.TypeName)
			continue
		}
		results = append(results, MarkerFunctionCall{
			Pkg:          pkg,
			Function:     MarkerFunction(id.TypeName),
			Arguments:    ce.Args,
			TypeArgument: parseTypeArguments(ce.Fun, file.Imports),
			File:         file,
			Position:     pkg.Fset.Position(e.Pos()),
		})
	}
	return results
}

func parseFuncFromExpr(e ast.Expr, importMap importmap.ImportMap) TypeID {
	var (
		ok     bool
		typeID TypeID
	)
	switch t := e.(type) {
	case *ast.SelectorExpr:
		var xIdent *ast.Ident
		xIdent, ok = t.X.(*ast.Ident)
		if !ok {
			return typeID
		}
		typeID.PkgPath, _ = importMap.GetPackageForPrefix(xIdent.Name)
		typeID.TypeName = t.Sel.Name
		return typeID
	case *ast.IndexExpr:
		return parseFuncFromExpr(t.X, importMap)
	case *ast.Ident:
		typeID.DeclaredLocally = true
		typeID.TypeName = t.Name
		return typeID
	case *ast.StarExpr:
		typeID = parseFuncFromExpr(t.X, importMap)
		typeID.Indirection = Pointer
		return typeID
	}
	return TypeID{}
}

func parseTypeArguments(e ast.Expr, importMap importmap.ImportMap) *TypeID {
	var expr ast.Expr
	if idxExpr, ok := e.(*ast.IndexExpr); ok {
		expr = idxExpr.Index
	} else {
		return nil
	}
	typeID := parseFuncFromExpr(expr, importMap)

	return &typeID
}

func (m MarkerFunctionCall) ParseTypesFromArgs(foo ...bool) ([]TypeID, error) {
	var p bool
	if len(foo) > 0 {
		p = foo[0]
	}
	return parseFuncCallForTypes(m.Arguments, m.File.Imports, m.Pkg.Fset, p)
}

func (m MarkerFunctionCall) ParseSchemaMethod() (SchemaMethod, error) {
	if len(m.Arguments) != 1 {
		err := fmt.Errorf("Schema Method expects one argument but got %d, at %s", len(m.Arguments), m.Position)
		return SchemaMethod{}, err
	}
	switch expr := m.Arguments[0].(type) {
	default:
		fmt.Printf("%T %#v", expr, expr)
	}
	return SchemaMethod{}, nil
}

func parseFuncCallForTypes(args []ast.Expr, importMap importmap.ImportMap, fset *token.FileSet, p bool) ([]TypeID, error) {
	var results []TypeID

	for _, arg := range args {
		pos := fset.Position(arg.Pos())
		if typeID, err := parseLitForType(arg, importMap); err != nil {
			pos = fset.Position(arg.Pos())
			return nil, fmt.Errorf("unsupported arg at %s: %w", pos, err)
		} else {
			results = append(results, typeID)
		}
	}

	return results, nil
}

func parseLitForType(expr ast.Expr, importMap importmap.ImportMap) (TypeID, error) {
	switch t := expr.(type) {
	case *ast.CompositeLit:
		return parseFuncFromExpr(t.Type, importMap), nil
	case *ast.UnaryExpr:
		if t.Op != token.AND {
			return TypeID{}, errors.New("unary expression op must be &")
		}
		lit, ok := t.X.(*ast.CompositeLit)
		if !ok {
			return TypeID{}, fmt.Errorf("unary expression type expects composite literal but was %T", t.X)
		}
		answer := parseFuncFromExpr(lit.Type, importMap)
		answer.Indirection = Pointer

		return answer, nil
	case *ast.CallExpr:
		p, ok := t.Fun.(*ast.ParenExpr)
		if !ok {
			return TypeID{}, fmt.Errorf("CallExpr fun must be ParenExpr, got %T", t.Fun)
		}
		return parseFuncFromExpr(p.X, importMap), nil
	default:
		fmt.Printf("Unrecognized -- %T %#v\n", expr, expr)
		return TypeID{}, fmt.Errorf("Unrecognized -- %T %#v\n", expr, expr)
	}
}
