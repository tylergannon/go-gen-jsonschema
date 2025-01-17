package syntax

import (
	"errors"
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"go/token"
	"slices"
	"strings"
)

const (
	MarkerFuncNewJSONSchemaBuilder = "NewJSONSchemaBuilder" // NewJSONSchemaBuilder
	MarkerFuncNewJSONSchemaMethod  = "NewJSONSchemaMethod"  // NewJSONSchemaMethod
	MarkerFuncNewInterfaceImpl     = "NewInterfaceImpl"     // NewInterfaceImpl
	MarkerFuncNewEnumType          = "NewEnumType"          // NewEnumType
)

// TypeID is our structured representation of a type. It can represent named types,
// pointers, slices, arrays, and generics—plus marks invalid or disallowed types.
type (

	// MarkerFunctionCall denotes a call to one of the marker functions, found
	// in the scanned source code.
	MarkerFunctionCall struct {
		CallExpr CallExpr
		// Our function calls need either zero or one type argument.
		// If present, denote the type argument here.
		TypeArgument *TypeID
		Arguments    []dst.Expr
	}
)

func (m MarkerFunctionCall) String() string {
	args := make([]string, len(m.Arguments))
	for i, arg := range m.Arguments {
		args[i] = fmt.Sprint(arg)
	}
	return fmt.Sprintf("%s %s Args{%s}", m.CallExpr.MustIdentifyFunc(), m.TypeArgument, strings.Join(args, ","))
}

var markerFunctions = []string{
	MarkerFuncNewJSONSchemaBuilder,
	MarkerFuncNewJSONSchemaMethod,
	MarkerFuncNewInterfaceImpl,
	MarkerFuncNewEnumType,
}

func ParseValueExprForMarkerFunctionCall(e ValueSpec) []MarkerFunctionCall {
	var results []MarkerFunctionCall
	for _, arg := range e.Value().Values {
		ce, ok := arg.(*dst.CallExpr)
		if !ok {
			continue
		}
		callExpr := NewCallExpr(ce, e.pkg, e.file)

		if id, ok := callExpr.IdentifyFunc(); !ok || id.PkgPath != SchemaPackagePath {
			fmt.Println("Not path", id)
			continue
		} else if !slices.Contains(markerFunctions, id.TypeName) {
			fmt.Println("Unsupported MarkerFunction", id.TypeName)
			continue
		}
		results = append(results, MarkerFunctionCall{
			CallExpr:     callExpr,
			Arguments:    ce.Args,
			TypeArgument: parseTypeArguments(ce.Fun, e.pkg, e.file.Imports),
		})
	}
	return results
}

func parseFuncFromExpr(e Expr) TypeID {
	var (
		ok     bool
		typeID TypeID
	)
	switch t := e.Expr().(type) {
	case *dst.SelectorExpr:
		var xIdent *dst.Ident
		xIdent, ok = t.X.(*dst.Ident)
		if !ok {
			return typeID
		}
		typeID.PkgPath, _ = e.Imports().GetPackageForPrefix(xIdent.Name)
		typeID.TypeName = t.Sel.Name
		return typeID
	case *dst.IndexExpr:
		return parseFuncFromExpr(e.NewExpr(t.X))
	case *dst.Ident:
		if t.Path == "" {
			typeID.PkgPath = e.Pkg().PkgPath
		} else {
			typeID.PkgPath = t.Path
		}
		typeID.TypeName = t.Name
		return typeID
	case *dst.StarExpr:
		typeID = parseFuncFromExpr(e.NewExpr(t.X))
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
	typeID := parseFuncFromExpr(NewExpr(expr, pkg, nil))

	return &typeID
}

func (m MarkerFunctionCall) ParseTypesFromArgs(foo ...bool) ([]TypeID, error) {
	var p bool
	if len(foo) > 0 {
		p = foo[0]
	}
	// TODO: Fix this one
	return parseFuncCallForTypes(m.Arguments, m.CallExpr.File(), m.CallExpr.Pkg(), p)
}

func unwrapSchemaMethodReceiver(expr Expr) (TypeID, error) {
	switch t := expr.Expr().(type) {
	case *dst.Ident:
		return TypeID{PkgPath: expr.Pkg().PkgPath, TypeName: t.Name}, nil
	case *dst.SelectorExpr:
		xIdent, ok := t.X.(*dst.Ident)
		if !ok {
			pos := expr.NewExpr(t.X).Position()
			return TypeID{}, fmt.Errorf("expected identifier, got (%T) %s at %s", t.X, t.X, pos)
		}
		pkgPath, ok := expr.Imports().GetPackageForPrefix(xIdent.Name)
		if !ok {
			pos := expr.NewExpr(t.X).Position()
			return TypeID{}, fmt.Errorf("couldn't find package for %s at %s", xIdent.Name, pos)
		}
		return TypeID{PkgPath: pkgPath, TypeName: t.Sel.Name}, nil
	case *dst.ParenExpr:
		return unwrapSchemaMethodReceiver(expr.NewExpr(t.X))
	case *dst.StarExpr:
		typeID, err := unwrapSchemaMethodReceiver(expr.NewExpr(t.X))
		if err != nil {
			return TypeID{}, err
		}
		typeID.Indirection = Pointer
		return typeID, nil
	default:
		return TypeID{}, fmt.Errorf("unrecognized schema method receiver expression at %s", expr.Position())
	}
}

func (m MarkerFunctionCall) ParseSchemaFunc() (SchemaFunction, error) {
	if m.TypeArgument == nil {
		return SchemaFunction{}, fmt.Errorf("expected a type argument to denote schema func at %s", m.CallExpr.Position())
	}
	return SchemaFunction{
		MarkerCall: m,
		Receiver:   *m.TypeArgument,
		FuncName:   m.Arguments[0].(*dst.Ident).Name,
	}, nil
}

func (m MarkerFunctionCall) ParseSchemaMethod() (SchemaMethod, error) {
	if len(m.Arguments) != 1 {
		err := fmt.Errorf("schema Method expects one argument but got %d, at %s", len(m.Arguments), m.CallExpr.Position())
		return SchemaMethod{}, err
	}
	switch expr := m.Arguments[0].(type) {
	// Must be a selector expression, in which X is either an Ident or a ParenExpr with a StarExpr to an Ident.
	case *dst.SelectorExpr:
		receiver, err := unwrapSchemaMethodReceiver(NewExpr(expr.X, m.CallExpr.pkg, m.CallExpr.file))
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

func parseFuncCallForTypes(args []dst.Expr, file *dst.File, pkg *decorator.Package, p bool) ([]TypeID, error) {
	var results []TypeID

	for _, arg := range args {

		pos := pkg.Fset.Position(pkg.Decorator.Map.Ast.Nodes[arg].Pos())
		if typeID, err := parseLitForType(NewExpr(arg, pkg, file)); err != nil {
			return nil, fmt.Errorf("unsupported arg at %s: %w", pos, err)
		} else {
			results = append(results, typeID)
		}
	}

	return results, nil
}

func parseLitForType(expr Expr) (TypeID, error) {
	switch t := expr.Expr().(type) {
	case *dst.CompositeLit:
		return parseFuncFromExpr(expr.NewExpr(t.Type)), nil
	case *dst.UnaryExpr:
		if t.Op != token.AND {
			return TypeID{}, errors.New("unary expression op must be &")
		}
		lit, ok := t.X.(*dst.CompositeLit)
		if !ok {
			return TypeID{}, fmt.Errorf("unary expression type expects composite literal but was %T", t.X)
		}
		answer := parseFuncFromExpr(expr.NewExpr(lit.Type))
		answer.Indirection = Pointer

		return answer, nil
	case *dst.CallExpr:
		p, ok := t.Fun.(*dst.ParenExpr)
		if !ok {
			return TypeID{}, fmt.Errorf("CallExpr fun must be ParenExpr, got %T", t.Fun)
		}
		return parseFuncFromExpr(expr.NewExpr(p.X)), nil
	default:
		fmt.Printf("Unrecognized -- %T %#v\n", expr, expr)
		return TypeID{}, fmt.Errorf("Unrecognized -- %T %#v\n", expr, expr)
	}
}
