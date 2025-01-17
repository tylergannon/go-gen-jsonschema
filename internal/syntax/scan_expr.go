package syntax

import (
	"errors"
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"go/token"
	"strings"
)

type (
	// Indirection labels a TypeID to tell whether the indicated type is a
	// concrete instance of a named type, a pointer to it, etc.

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

// TypeID is our structured representation of a type. It can represent named types,
// pointers, slices, arrays, and generics—plus marks invalid or disallowed types.
type (

	// MarkerFunctionCall denotes a call to one of the marker functions, found
	// in the scanned source code.
	MarkerFunctionCall struct {
		Pkg      *decorator.Package
		Function MarkerFunction
		// Our function calls need either zero or one type argument.
		// If present, denote the type argument here.
		TypeArgument *TypeID
		Arguments    []dst.Expr
		File         *dst.File
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

func ParseValueExprForMarkerFunctionCall(e ValueSpec) []MarkerFunctionCall {
	var results []MarkerFunctionCall
	for _, arg := range e.Value().Values {
		ce, ok := arg.(*dst.CallExpr)
		if !ok {
			continue
		}
		id := parseFuncFromExpr(ce.Fun, e.Imports(), e.Pkg())
		if id.PkgPath != SchemaPackagePath {
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
			Pkg:          e.pkg,
			Function:     MarkerFunction(id.TypeName),
			Arguments:    ce.Args,
			TypeArgument: parseTypeArguments(ce.Fun, e.pkg, e.file.Imports),
			File:         e.file,
			Position:     e.Position(),
		})
	}
	return results
}

func parseFuncFromExpr(e dst.Expr, importMap ImportMap, pkg *decorator.Package) TypeID {
	var (
		ok     bool
		typeID TypeID
	)
	switch t := e.(type) {
	case *dst.SelectorExpr:
		var xIdent *dst.Ident
		xIdent, ok = t.X.(*dst.Ident)
		if !ok {
			return typeID
		}
		typeID.PkgPath, _ = importMap.GetPackageForPrefix(xIdent.Name)
		typeID.TypeName = t.Sel.Name
		return typeID
	case *dst.IndexExpr:
		return parseFuncFromExpr(t.X, importMap, pkg)
	case *dst.Ident:
		if t.Path == "" {
			typeID.PkgPath = pkg.PkgPath
		} else {
			typeID.PkgPath = t.Path
		}
		typeID.TypeName = t.Name
		return typeID
	case *dst.StarExpr:
		typeID = parseFuncFromExpr(t.X, importMap, pkg)
		typeID.Indirection = Pointer
		return typeID
	}
	return TypeID{}
}

func parseTypeArguments(e dst.Expr, pkg *decorator.Package, importMap ImportMap) *TypeID {
	var expr dst.Expr
	if idxExpr, ok := e.(*dst.IndexExpr); ok {
		expr = idxExpr.Index
	} else {
		return nil
	}
	typeID := parseFuncFromExpr(expr, importMap, pkg)

	return &typeID
}

func (m MarkerFunctionCall) ParseTypesFromArgs(foo ...bool) ([]TypeID, error) {
	var p bool
	if len(foo) > 0 {
		p = foo[0]
	}
	return parseFuncCallForTypes(m.Arguments, m.File.Imports, m.Pkg, p)
}

func unwrapSchemaMethodReceiver(expr dst.Expr, pkg *decorator.Package, importMap ImportMap) (TypeID, error) {
	switch t := expr.(type) {
	case *dst.Ident:
		return TypeID{PkgPath: pkg.PkgPath, TypeName: t.Name}, nil
	case *dst.SelectorExpr:
		xIdent, ok := t.X.(*dst.Ident)
		if !ok {

			pos := pkg.Fset.Position(pkg.Decorator.Map.Ast.Nodes[t.X].Pos())
			return TypeID{}, fmt.Errorf("expected identifier, got (%T) %s at %s", t.X, t.X, pos.String())
		}
		pkgPath, ok := importMap.GetPackageForPrefix(xIdent.Name)
		if !ok {
			pos := pkg.Fset.Position(pkg.Decorator.Map.Ast.Nodes[t.X].Pos())
			return TypeID{}, fmt.Errorf("couldn't find package for %s at %s", xIdent.Name, pos)
		}
		return TypeID{PkgPath: pkgPath, TypeName: t.Sel.Name}, nil
	case *dst.ParenExpr:
		return unwrapSchemaMethodReceiver(t.X, pkg, importMap)
	case *dst.StarExpr:
		typeID, err := unwrapSchemaMethodReceiver(t.X, pkg, importMap)
		if err != nil {
			return TypeID{}, err
		}
		typeID.Indirection = Pointer
		return typeID, nil
	default:
		pos := NodePosition(pkg, t)
		return TypeID{}, fmt.Errorf("unrecognized schema method receiver expression at %s", pos)
	}
}

func nodePos(pkg *decorator.Package, node dst.Node) token.Pos {
	return pkg.Decorator.Map.Ast.Nodes[node].Pos()
}

func NodePosition(pkg *decorator.Package, node dst.Node) token.Position {
	return pkg.Fset.Position(nodePos(pkg, node))
}

func (m MarkerFunctionCall) ParseSchemaFunc() (SchemaFunction, error) {
	if m.TypeArgument == nil {
		return SchemaFunction{}, fmt.Errorf("expected a type argument to denote schema func at %s", m.Position)
	}
	return SchemaFunction{
		MarkerCall: m,
		Receiver:   *m.TypeArgument,
		FuncName:   m.Arguments[0].(*dst.Ident).Name,
	}, nil
}

func (m MarkerFunctionCall) ParseSchemaMethod() (SchemaMethod, error) {
	if len(m.Arguments) != 1 {
		err := fmt.Errorf("schema Method expects one argument but got %d, at %s", len(m.Arguments), m.Position)
		return SchemaMethod{}, err
	}
	switch expr := m.Arguments[0].(type) {
	// Must be a selector expression, in which X is either an Ident or a ParenExpr with a StarExpr to an Ident.
	case *dst.SelectorExpr:
		receiver, err := unwrapSchemaMethodReceiver(expr.X, m.Pkg, m.File.Imports)
		if err != nil {
			return SchemaMethod{}, err
		}
		return SchemaMethod{
			Receiver:   receiver,
			FuncName:   expr.Sel.Name,
			MarkerCall: m,
		}, nil
	default:
		fmt.Printf("ArgBoo --> %T %#v", expr, expr)
	}
	return SchemaMethod{}, nil
}

func parseFuncCallForTypes(args []dst.Expr, importMap ImportMap, pkg *decorator.Package, p bool) ([]TypeID, error) {
	var results []TypeID

	for _, arg := range args {

		pos := pkg.Fset.Position(pkg.Decorator.Map.Ast.Nodes[arg].Pos())
		if typeID, err := parseLitForType(arg, pkg, importMap); err != nil {
			return nil, fmt.Errorf("unsupported arg at %s: %w", pos, err)
		} else {
			results = append(results, typeID)
		}
	}

	return results, nil
}

func parseLitForType(expr dst.Expr, pkg *decorator.Package, importMap ImportMap) (TypeID, error) {
	switch t := expr.(type) {
	case *dst.CompositeLit:
		return parseFuncFromExpr(t.Type, importMap, pkg), nil
	case *dst.UnaryExpr:
		if t.Op != token.AND {
			return TypeID{}, errors.New("unary expression op must be &")
		}
		lit, ok := t.X.(*dst.CompositeLit)
		if !ok {
			return TypeID{}, fmt.Errorf("unary expression type expects composite literal but was %T", t.X)
		}
		answer := parseFuncFromExpr(lit.Type, importMap, pkg)
		answer.Indirection = Pointer

		return answer, nil
	case *dst.CallExpr:
		p, ok := t.Fun.(*dst.ParenExpr)
		if !ok {
			return TypeID{}, fmt.Errorf("CallExpr fun must be ParenExpr, got %T", t.Fun)
		}
		return parseFuncFromExpr(p.X, importMap, pkg), nil
	default:
		fmt.Printf("Unrecognized -- %T %#v\n", expr, expr)
		return TypeID{}, fmt.Errorf("Unrecognized -- %T %#v\n", expr, expr)
	}
}
