package syntax

import (
	"errors"
	"fmt"
	"go/token"
	"slices"
	"strings"

	"github.com/dave/dst"
)

const (
	MarkerFuncNewJSONSchemaBuilder = "NewJSONSchemaBuilder" // NewJSONSchemaBuilder
	MarkerFuncNewJSONSchemaMethod  = "NewJSONSchemaMethod"  // NewJSONSchemaMethod
	MarkerFuncNewJSONSchemaFunc    = "NewJSONSchemaFunc"    // NewJSONSchemaFunc
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
	}
)

func (m MarkerFunctionCall) MustTypeArgument() TypeID {
	if t := m.TypeArgument(); t == nil {
		panic("type argument cannot be nil")
	} else {
		return *t
	}
}

func (m MarkerFunctionCall) TypeArgument() *TypeID {
	var expr dst.Expr
	if idxExpr, ok := m.CallExpr.Concrete.Fun.(*dst.IndexExpr); ok {
		expr = idxExpr.Index
	} else {
		return nil
	}
	typeID := parseFuncFromExpr(m.CallExpr.NewExpr(expr))
	return &typeID
}

func (m MarkerFunctionCall) String() string {
	callArgs := m.CallExpr.Args()
	args := make([]string, len(callArgs))
	for i, arg := range callArgs {
		args[i] = fmt.Sprint(arg)
	}
	return fmt.Sprintf("%s %s Args{%s}", m.CallExpr.MustIdentifyFunc(), m.TypeArgument(), strings.Join(args, ","))
}

var markerFunctions = []string{
	MarkerFuncNewJSONSchemaBuilder,
	MarkerFuncNewJSONSchemaMethod,
	MarkerFuncNewJSONSchemaFunc,
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
			continue
		} else if !slices.Contains(markerFunctions, id.TypeName) {
			fmt.Println("Unsupported MarkerFunction", id.TypeName)
			continue
		}
		results = append(results, MarkerFunctionCall{
			CallExpr: callExpr,
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

//func parseTypeArguments(e dst.Field, pkgPath *decorator.Package, importMap ImportMap) *TypeID {
//	var expr dst.Field
//	if idxExpr, ok := e.(*dst.IndexExpr); ok {
//		expr = idxExpr.Index
//	} else {
//		return nil
//	}
//	typeID := parseFuncFromExpr(NewExpr(expr, pkgPath, nil))
//
//	return &typeID
//}

func (m MarkerFunctionCall) ParseTypesFromArgs() ([]TypeID, error) {
	var results []TypeID
	for _, arg := range m.CallExpr.Args() {
		if typeID, err := parseLitForType(arg); err != nil {
			return nil, fmt.Errorf("unsupported arg at %s: %w", arg.Position(), err)
		} else {
			results = append(results, typeID)
		}
	}

	return results, nil
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
	var typeArg = m.TypeArgument()
	if typeArg == nil {
		return SchemaFunction{}, fmt.Errorf("expected a type argument to denote schema func at %s", m.CallExpr.Position())
	}
	// Build base schema function from type arg and first argument (func ident)
	sf := SchemaFunction{
		MarkerCall:       m,
		Receiver:         *typeArg,
		SchemaMethodName: m.CallExpr.Args()[0].Expr().(*dst.Ident).Name,
	}
	// Parse optional consolidated options (variadic args beyond the first)
	if len(m.CallExpr.Args()) > 1 {
		if opts, err := parseSchemaMethodOptions(m.CallExpr.Args()[1:], *typeArg, m); err == nil {
			sf.Options = opts
		}
	}
	return sf, nil
}

// ParseSchemaBuilder supports two forms:
//  1. Back-compat: NewJSONSchemaBuilder[T](Func)
//  2. v1 form:     NewJSONSchemaBuilder(Func, options...) where T is inferred
//     from the first consolidated option that references a field via
//     ExampleType{}.Field. When no such option is present, returns an error.
func (m MarkerFunctionCall) ParseSchemaBuilder() (SchemaFunction, error) {
	args := m.CallExpr.Args()
	if len(args) < 1 {
		return SchemaFunction{}, fmt.Errorf("schema Builder expects at least one argument (the func), at %s", m.CallExpr.Position())
	}
	// Prefer explicit type argument when present (back-compat)
	if ta := m.TypeArgument(); ta != nil {
		return SchemaFunction{
			MarkerCall:       m,
			Receiver:         *ta,
			SchemaMethodName: args[0].Expr().(*dst.Ident).Name,
		}, nil
	}
	// Infer from options: scan variadic args beyond the first for With* calls
	var inferred TypeID
	var options []SchemaMethodOptionInfo
	if len(args) > 1 {
		for _, a := range args[1:] {
			ce, ok := a.Expr().(*dst.CallExpr)
			if !ok {
				continue
			}
			funID := parseFuncFromExpr(a.NewExpr(ce.Fun))
			if funID.PkgPath != SchemaPackagePath {
				continue
			}
			if len(ce.Args) != 2 {
				continue
			}
			// Expect first arg as ExampleType{}.Field
			fieldSel, ok := ce.Args[0].(*dst.SelectorExpr)
			if !ok {
				continue
			}
			lit, ok := fieldSel.X.(*dst.CompositeLit)
			if !ok {
				continue
			}
			if id, ok := lit.Type.(*dst.Ident); ok {
				if inferred.TypeName == "" {
					inferred = TypeID{PkgPath: m.CallExpr.pkg.PkgPath, TypeName: id.Name}
				}
			}
			// Collect option metadata similar to parseSchemaMethodOptions, but without receiver filtering
			var providerName string
			providerIsMethod := false
			switch p := ce.Args[1].(type) {
			case *dst.SelectorExpr:
				providerIsMethod = true
				providerName = p.Sel.Name
			case *dst.Ident:
				providerName = p.Name
			}
			var kind SchemaMethodOptionKind
			funName := funID.TypeName
			switch funName {
			case "WithFunction":
				kind = SchemaMethodOptionKind("WithFunction")
			case "WithStructAccessorMethod":
				kind = SchemaMethodOptionKind("WithStructAccessorMethod")
			case "WithStructFunctionMethod":
				kind = SchemaMethodOptionKind("WithStructFunctionMethod")
			default:
				// Unknown option kinds ignored for inference
			}
			if kind != "" {
				options = append(options, SchemaMethodOptionInfo{
					Kind:             kind,
					FieldName:        fieldSel.Sel.Name,
					ProviderName:     providerName,
					ProviderIsMethod: providerIsMethod,
				})
			}
		}
	}
	if inferred.TypeName == "" {
		return SchemaFunction{}, fmt.Errorf("unable to infer receiver type for NewJSONSchemaBuilder at %s: add a consolidated option like WithFunction(T{}.Field, Provider)", m.CallExpr.Position())
	}
	return SchemaFunction{
		MarkerCall:       m,
		Receiver:         inferred,
		SchemaMethodName: args[0].Expr().(*dst.Ident).Name,
		Options:          options,
	}, nil
}

func (m MarkerFunctionCall) Args() []Expr {
	return m.CallExpr.Args()
}

func parseSchemaMethodOptions(args []Expr, receiver TypeID, m MarkerFunctionCall) ([]SchemaMethodOptionInfo, error) {
	var out []SchemaMethodOptionInfo
	for _, a := range args {
		ce, ok := a.Expr().(*dst.CallExpr)
		if !ok {
			continue
		}
		funID := parseFuncFromExpr(a.NewExpr(ce.Fun))
		if funID.PkgPath != SchemaPackagePath {
			continue
		}
		// Special-case zero-arg option WithRenderProviders()
		if funID.TypeName == "WithRenderProviders" {
			out = append(out, SchemaMethodOptionInfo{Kind: SchemaMethodOptionKind("WithRenderProviders")})
			continue
		}
		// WithDiscriminator(field, "name") has 2 args, second is string literal
		if funID.TypeName == "WithDiscriminator" && len(ce.Args) == 2 {
			fieldSel, ok := ce.Args[0].(*dst.SelectorExpr)
			if !ok {
				continue
			}
			lit, ok := fieldSel.X.(*dst.CompositeLit)
			if !ok {
				continue
			}
			recvIdent, ok := lit.Type.(*dst.Ident)
			if !ok || recvIdent.Name != receiver.TypeName {
				continue
			}
			if str, ok := ce.Args[1].(*dst.BasicLit); ok && str.Kind == token.STRING {
				name := strings.Trim(str.Value, "\"")
				out = append(out, SchemaMethodOptionInfo{Kind: SchemaMethodOptionKind("WithDiscriminator"), FieldName: fieldSel.Sel.Name, Discriminator: name})
			}
			continue
		}
		if len(ce.Args) < 1 {
			continue
		}
		// First arg: exampleStruct{}.FieldX
		fieldSel, ok := ce.Args[0].(*dst.SelectorExpr)

		if !ok {
			continue
		}
		lit, ok := fieldSel.X.(*dst.CompositeLit)
		if !ok {
			continue
		}
		recvIdent, ok := lit.Type.(*dst.Ident)
		if !ok || recvIdent.Name != receiver.TypeName {
			continue
		}
		fieldName := fieldSel.Sel.Name
		var providerName string
		providerIsMethod := false
		var kind SchemaMethodOptionKind
		funName := funID.TypeName
		switch funName {
		case "WithFunction":
			kind = SchemaMethodOptionKind("WithFunction")
		case "WithStructAccessorMethod":
			kind = SchemaMethodOptionKind("WithStructAccessorMethod")
		case "WithStructFunctionMethod":
			kind = SchemaMethodOptionKind("WithStructFunctionMethod")
		case "WithInterface":
			kind = SchemaMethodOptionKind("WithInterface")
		case "WithInterfaceImpls":
			kind = SchemaMethodOptionKind("WithInterfaceImpls")
		case "WithEnum":
			kind = SchemaMethodOptionKind("WithEnum")
		case "WithStringerEnum":
			kind = SchemaMethodOptionKind("WithStringerEnum")
		default:
			continue
		}
		// Only parse provider if applicable
		if (funName == "WithFunction" || funName == "WithStructAccessorMethod" || funName == "WithStructFunctionMethod") && len(ce.Args) > 1 {
			provExpr := ce.Args[1]
			switch p := provExpr.(type) {
			case *dst.SelectorExpr:
				// Expect ReceiverType.MethodName or (ReceiverType).MethodName
				switch x := p.X.(type) {
				case *dst.Ident:
					if x.Name != receiver.TypeName {
						continue
					}
					providerIsMethod = true
					providerName = p.Sel.Name
				case *dst.ParenExpr:
					if id, ok := x.X.(*dst.Ident); ok && id.Name == receiver.TypeName {
						providerIsMethod = true
						providerName = p.Sel.Name
					} else {
						continue
					}
				default:
					continue
				}
			case *dst.Ident:
				providerName = p.Name
			}
		}
		var impls []TypeID
		if funID.TypeName == "WithInterfaceImpls" && len(ce.Args) > 1 {
			for _, a2 := range ce.Args[1:] {
				if tid, err := parseLitForType(NewExpr(a2, m.CallExpr.pkg, m.CallExpr.file)); err == nil {
					impls = append(impls, tid)
				}
			}
		}
		out = append(out, SchemaMethodOptionInfo{
			Kind:             kind,
			FieldName:        fieldName,
			ProviderName:     providerName,
			ProviderIsMethod: providerIsMethod,
			ImplTypes:        impls,
		})
	}
	return out, nil
}

func (m MarkerFunctionCall) ParseSchemaMethod() (SchemaMethod, error) {
	// There's only one result object because we only accept a single
	// argument to the NewJSONSchema method.
	var funcArgs = m.CallExpr.Args()
	if len(funcArgs) < 1 {
		return SchemaMethod{}, fmt.Errorf("schema Method expects at least one argument (the method), at %s", m.CallExpr.Position())
	}
	switch expr := funcArgs[0].Expr().(type) {
	// Must be a selector expression, in which X is either an Ident or a ParenExpr with a StarExpr to an Ident.
	case *dst.SelectorExpr:
		receiver, err := unwrapSchemaMethodReceiver(NewExpr(expr.X, m.CallExpr.pkg, m.CallExpr.file))
		if err != nil {
			return SchemaMethod{}, err
		}
		res := SchemaMethod{
			Receiver:         receiver,
			SchemaMethodName: expr.Sel.Name,
			MarkerCall:       m,
		}
		// Parse optional sentinel options (variadic args beyond the first)
		if len(funcArgs) > 1 {
			if opts, err := parseSchemaMethodOptions(funcArgs[1:], receiver, m); err == nil {
				res.Options = opts
			}
		}
		return res, nil
	default:
		fmt.Printf("ArgBoo --> %T %#v", expr, expr)
	}
	return SchemaMethod{}, nil
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
		return TypeID{}, fmt.Errorf("unrecognized -- %T %#v", expr, expr)
	}
}
