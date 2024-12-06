package typeregistry

import (
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/tylergannon/structtag"
	"go/token"
	"go/types"
	"strings"
)

type edgeType int

const (
	edgeTypeIdentity edgeType = iota
	edgeTypeField
	edgeTypeDefinition
)

type edge struct {
	Type edgeType
	From TypeID
	To   TypeID
}

type nodeState int

const (
	nodeStateNew nodeState = iota
	nodeStateVisited
	nodeStateProcessed
)

type Node struct {
	parent   *Node              // the Node that caused this to go onto the stack
	typ      types.Type         // the types object
	expr     dst.Node           // should be either a TypeSpec or a Expr, depending on whether the typ is a named type or not.
	pkg      *decorator.Package // the decorated package that this Node occurs in
	typeSpec TypeSpec           // may be nil; only exists for top-level type decls.
	children []*edge            // before leaving from the first visit, must be initialized to have length equal to number of child nodes
	// The following are determined on the second visit.
	id TypeID // update this after resolving the type
	//indexAtParent int                // the index where this Node should reflect its resolved type within `parent.children`.
	state nodeState
}

func (r *Registry) graphTypeForSchema(ts TypeSpec) (*Node, map[TypeID]*Node, error) {
	var (
		nodes    = map[TypeID]*Node{}
		rootNode = &Node{
			typ:      ts.GetType(),
			typeSpec: ts,
			pkg:      ts.Pkg(),
			expr:     ts.GetTypeSpec(),
		}
		stack = []*Node{rootNode}
	)

	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if nodesTemp, err := r.visitNode(node, nodes); err != nil {
			return nil, nil, err
		} else {
			stack = append(stack, nodesTemp...)
		}
	}
	return rootNode, nodes, nil
}

func (r *Registry) resolveTypeIdent(ident *dst.Ident, currPkgPath *decorator.Package) (TypeSpec, error) {
	return nil, nil
}

// visitNode does the first pass on a given node.
// Basically finds all direct children according to the AST, and returns them
// as a slice of new nodes.
// If the type of the object is a struct type, there will be more than one
// child node.
// (Later) if the node is a "type alternatives" node, there may be more than
// one node.
// The default case is that
func (r *Registry) visitNode(node *Node, nodes map[TypeID]*Node) ([]*Node, error) {
	switch node.typ.(type) {
	case *types.Alias:
		return nil, fmt.Errorf("alias types not yet supported: %w", ErrUnsupportedType)
	case *types.Named:
		return r.visitNamedTypeNode(node, nodes)
	case *types.Struct:
		return r.visitStructTypeNode(node, nodes)
	case *types.Pointer:
		panic("we don't do pointers here")

	default:
		inspect("expr", node.expr)
		inspect("typ", node.typ)
	}

	return nil, nil
}

func (r *Registry) visitStructTypeNode(node *Node, nodesMap map[TypeID]*Node) ([]*Node, error) {
	var (
		structTyp  = assertType[*types.Struct](node.typ)
		structDecl = assertType[*dst.StructType](node.expr)
		nodes      []*Node
	)
	for i := 0; i < structTyp.NumFields(); i++ {
		var (
			field     = structTyp.Field(i)
			fieldDecl = structDecl.Fields.List[i]
			conf, err = parseFieldConf(field, fieldDecl, node.pkg)
			typeID    TypeID
		)
		if err != nil {
			return nil, err
		}
		if conf.ignore() {
			continue
		}
		if typeID, err = r.resolveType(field.Type(), fieldDecl.Type, node.pkg); err != nil {
			return nil, err
		}
		nodes = append(nodes, &Node{
			parent:   node,
			typ:      field.Type(),
			expr:     fieldDecl.Type,
			pkg:      node.pkg,
			typeSpec: nil,
			children: nil,
			id:       typeID,
			state:    0,
		})

	}
	return nodes, nil
}

// visitNamedTypeNode
//  1. Find the AST node for the underlying type
//     (a) here
//     (b) just on the other side of a pointer
//     (c) at the end of a trail of name / definition indirection.
//  2. TODO: detect a custom `json.Marshaler` implementation, which is tricky
//     because in some cases we will emit such a thing.
//
// type, skipping over pointers,
func (r *Registry) visitNamedTypeNode(parent *Node, nodes map[TypeID]*Node) ([]*Node, error) {
	node := &Node{
		parent: parent,
		pkg:    parent.pkg, // must be overwritten when type is actually declared elsewhere.
	}
	namedType := parent.typ.(*types.Named)

	var typeSpecNode, ok = parent.expr.(*dst.TypeSpec)
	if !ok {
		return nil, fmt.Errorf("found a types.Named paired with dst.Node that's not a TypeSpec but %T: %w", parent.expr, ErrUnsupportedType)
	}
	var underlying, expr = r.dereferenceAndIdentify(namedType.Underlying(), typeSpecNode.Type, parent.pkg)

	switch t := expr.(type) {
	case *dst.Ident:
		switch underlying := underlying.(type) {
		case *types.Basic, *types.Named:
			node.typ = underlying
			node.expr = t
		}
		//if underlyingBasic, ok := underlying.(*types.Basic); ok {
		//	node.typ = underlyingBasic
		//	node.expr = t
		//} else {
		//	ts, _, ok := r.getType(t.Name, parent.pkg.PkgPath)
		//	if !ok {
		//		panic("I fucked up: " + t.Name)
		//	}
		//	inspect("ts expr", ts.typeSpec.Type)
		//}
	case *dst.StructType:
		node.typ = underlying
		node.expr = t
	case *dst.ArrayType:
		switch underlying := underlying.(type) {
		case *types.Slice, *types.Array:
			node.typ = underlying
		default:
			panic("should have gotten an array type here.")
		}
		node.expr = t
	case *dst.StarExpr:
		panic("Star Expr should not be possible here.")
	default:
		inspect("typeSpecNode is: ", t)

	}

	switch underlying.(type) {
	case *types.Pointer:
		panic("shouldn't get a pointer here.")
	}
	return []*Node{node}, nil
}

// If pointer type, dereference.
// If expr is an *dst.Ident, locate the type declaration.
func (r *Registry) dereferenceAndIdentify(typ types.Type, expr dst.Expr, localPkg *decorator.Package) (types.Type, dst.Expr) {
	switch _typ := typ.(type) {
	case *types.Basic:
		return typ, assertType[*dst.Ident](expr)
	case *types.Named:
		return typ, assertType[*dst.Ident](expr)
	case *types.Pointer:
		return r.dereferenceAndIdentify(_typ.Elem(), assertType[*dst.StarExpr](expr).X, localPkg)
	}

	// If typ is not a *types.Basic and expr is an ident, I think that means
	// typ should be a named type definition.
	// If expr is not ident, maybe we'll do a little pattern matching
	// just to assert types match.
	switch exprType := expr.(type) {
	case *dst.StarExpr:
		panic("got a star expr after already checking for pointer type")
	case *dst.Ident:
		ts, err := r.locateType(exprType, localPkg)
		if err != nil {
			panic(err)
		}
		return typ, ts.GetTypeSpec().Type
	}

	return typ, expr
}

func (r *Registry) locateType(ident *dst.Ident, localPkg *decorator.Package) (TypeSpec, error) {
	var path = ident.Path
	if path == "" {
		path = localPkg.PkgPath
	} else {
		if err := r.LoadAndScan(path); err != nil {
			return nil, fmt.Errorf("trying to locate type %v: %w", ident, err)
		}
	}
	ts, _, ok := r.getType(ident.Name, path)
	if !ok {
		return nil, fmt.Errorf("type not found trying to locate %v", ident)
	}
	_typeSpec := ts.typeSpec
	if _ident, ok := _typeSpec.Type.(*dst.Ident); ok {
		return r.locateType(_ident, ts.pkg)
	}
	return ts, nil
}

func assertType[T any](expr any) T {
	if _expr, ok := expr.(T); !ok {
		var x T
		panic(fmt.Sprintf("expected %v to be of type %T but was %T", expr, x, expr))
	} else {
		return _expr
	}
}

func assertNonPointer(typ types.Type, expr dst.Expr) (types.Type, dst.Expr) {
	if _, ok := typ.(*types.Pointer); ok {
		panic(fmt.Sprintf("type mismatch between %T and %T: %v", typ, expr, ErrUnsupportedType))
	}
	if _, ok := expr.(*dst.StarExpr); ok {
		panic(fmt.Sprintf("type mismatch between %T and %T: %v", typ, expr, ErrUnsupportedType))
	}
	return typ, expr
}

type structFieldConf struct {
	jsonTagOptions []string
	fieldName      string
	*dst.Field
	*types.Var
	pkg *decorator.Package
}

func (f structFieldConf) ignore() bool {
	return f.fieldName == "-" || (f.Tag == nil && !f.Var.Exported())
}

func (f structFieldConf) typeExpr() dst.Expr {
	return f.Field.Type
}

func (f structFieldConf) fieldType() types.Type {
	return f.Var.Type()
}

func (f structFieldConf) position() token.Position {
	return f.pkg.Fset.Position(f.Pos())
}

func (f structFieldConf) posString() string {
	p := f.position()
	return fmt.Sprintf("at %s:%d:%d", p.Filename, p.Line, p.Column)
}

func parseFieldConf(field *types.Var, fieldNode *dst.Field, pkg *decorator.Package) (structFieldConf, error) {
	var (
		fieldName      string
		jsonTagOptions []string
	)
	if !field.Embedded() && fieldNode.Tag != nil {
		tagValue := strings.Trim(fieldNode.Tag.Value, "`")
		if tags, err := structtag.Parse(tagValue); err != nil {
			position := pkg.Fset.Position(field.Pos())

			return structFieldConf{}, fmt.Errorf(
				"failed to parse tags for field %q at %s:%d:%d: %w",
				field.Name(), position.Filename, position.Line, position.Column, err,
			)
		} else if tag, err := tags.Get("json"); err != nil || len(tag.Options) == 0 {
			fieldName = field.Name()
		} else {
			fieldName = tag.Options[0]
			jsonTagOptions = tag.Options
		}
	} else if !field.Embedded() {
		fieldName = field.Name()
	}
	return structFieldConf{
		jsonTagOptions: jsonTagOptions,
		fieldName:      fieldName,
		Var:            field,
		Field:          fieldNode,
		pkg:            pkg,
	}, nil
}
