package typeregistry

import (
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/tylergannon/structtag"
	"go/token"
	"go/types"
	"runtime/debug"
	"strings"
)

type Node struct {
	//parent   *Node              // the Node that caused this to go onto the stack
	Type     types.Type         // the types object
	Node     dst.Node           // should be either a TypeSpec or a Expr, depending on whether the Type is a named type or not.
	Pkg      *decorator.Package // the decorated package that this Node occurs in
	TypeSpec TypeSpec           // may be nil; only exists for top-level type decls.
	Children []TypeID           // before leaving from the first visit, must be initialized to have length equal to number of child nodes
	inbound  int                // Number of edges with this node as destination
	// The following are determined on the second visit.
	ID TypeID // update this after resolving the type
	//indexAtParent int                // the index where this Node should reflect its resolved type within `parent.Children`.
	structField *StructFieldConf
}

func countInbound(nodes map[TypeID]*Node) map[TypeID]*Node {
	counts := make(map[TypeID]int, len(nodes))
	for _, n := range nodes {
		for _, child := range n.Children {
			counts[child]++
		}
	}
	for id, n := range nodes {
		n.inbound = counts[id]
	}
	return nodes
}

type SchemaGraph struct {
	RootNode *Node
	Nodes    map[TypeID]*Node
}

func (r *Registry) GraphTypeForSchema(ts TypeSpec) (*SchemaGraph, error) {
	var (
		nodes    = map[TypeID]*Node{}
		rootNode = &Node{
			Type:     ts.GetType(),
			TypeSpec: ts,
			Pkg:      ts.Pkg(),
			Node:     ts.GetTypeSpec(),
			ID:       NewTypeID(ts.Pkg().PkgPath, ts.GetType().(*types.Named).Obj().Name()),
		}
		stack = []*Node{rootNode}
	)

	for i := 0; len(stack) > 0; i++ {
		if i > 10000 {
			return nil, fmt.Errorf("iterated a little long, perhaps")
		}
		node := stack[0]
		stack = stack[1:]
		if _, ok := nodes[node.ID]; ok {
			continue
		}
		nodes[node.ID] = node

		if nodesTemp, err := r.visitNode(node); err != nil {
			return nil, err
		} else {
			for _, nodeTemp := range nodesTemp {
				node.Children = append(node.Children, nodeTemp.ID)
				if namedType, ok := nodeTemp.Type.(*types.Named); ok {
					if tsTemp, _, ok := r.getType(namedType.Obj().Name(), namedType.Obj().Pkg().Path()); !ok {
						panic("why I not find")
					} else {
						nodeTemp.TypeSpec = tsTemp
					}
				}
				if _, ok := nodes[nodeTemp.ID]; !ok {
					if len(nodeTemp.ID) == 0 {
						inspect("empty thing", nodeTemp.Type, nodeTemp.Node)
					}
					stack = append(stack, nodeTemp)
				}
			}
		}
	}
	return &SchemaGraph{RootNode: rootNode, Nodes: nodes}, nil
}

func (r *Registry) resolveTypeIdent(ident *dst.Ident, currPkgPath *decorator.Package) (TypeSpec, error) {
	pkgPath := ident.Path
	if ident.Path == "" {
		pkgPath = currPkgPath.PkgPath
	}
	ts, _, ok := r.getType(ident.Name, pkgPath)
	if !ok {
		return nil, fmt.Errorf("type identifier not found: %s", ident.Name)
	}
	return ts, nil
}

// visitNode does the first pass on a given node.
// Basically finds all direct Children according to the AST, and returns them
// as a slice of new nodes.
// If the type of the object is a struct type, there will be more than one
// child node.
// (Later) if the node is a "type alternatives" node, there may be more than
// one node.
// The default case is that
func (r *Registry) visitNode(node *Node) ([]*Node, error) {
	switch node.Type.(type) {
	case *types.Alias:
		return nil, fmt.Errorf("alias types not yet supported: %w", ErrUnsupportedType)
	case *types.Named:
		return r.visitNamedTypeNode(node)
	case *types.Struct:
		return r.visitStructTypeNode(node)
	case *types.Pointer:
		fmt.Println(string(debug.Stack()))
		panic("we don't do pointers here")
	case *types.Slice:
		return r.visitSliceTypeNode(node)
	case *types.Array:
		return r.visitArrayTypeNode(node)
	case *types.Basic:
		return nil, nil

	default:
		panic(fmt.Sprintf("unexpected node type: %T", node.Type))
		//inspect("Node", node.Node)
		//inspect("Type", node.Type)
	}

	return nil, nil
}

func (r *Registry) visitArrayTypeNode(node *Node) ([]*Node, error) {
	var (
		t           = node.Type.(*types.Array)
		ts          = node.Node.(*dst.ArrayType)
		typeID, err = r.resolveType(t.Elem(), ts.Elt, node.Pkg)
	)
	if err != nil {
		return nil, err
	}
	return []*Node{
		{
			Type:     t.Elem(),
			Node:     ts.Elt,
			Pkg:      node.Pkg,
			TypeSpec: nil,
			ID:       typeID,
		},
	}, nil
}
func (r *Registry) visitSliceTypeNode(node *Node) ([]*Node, error) {
	var (
		t           = node.Type.(*types.Slice)
		ts          = node.Node.(*dst.ArrayType)
		typeID, err = r.resolveType(t.Elem(), ts.Elt, node.Pkg)
	)
	if err != nil {
		return nil, err
	}
	return []*Node{
		{
			Type:     t.Elem(),
			Node:     ts.Elt,
			Pkg:      node.Pkg,
			TypeSpec: nil,
			ID:       typeID,
		},
	}, nil
}

func (r *Registry) visitEmbeddedStructField(parent *Node, field *types.Var) (nodes []*Node, err error) {
	if err = r.LoadAndScan(field.Pkg().Path()); err != nil {
		return nil, err
	}
	var (
		pkg        = r.packages[field.Pkg().Path()]
		pos        = pkg.Fset.Position(field.Pos())
		named      *types.Named
		ts         TypeSpec
		structType *types.Struct
		structNode *dst.StructType
		ok         bool
	)
	if named, ok = field.Type().(*types.Named); !ok {
		return nil, fmt.Errorf("embedded fields must be named type at %s:%d:%d", pos.Filename, pos.Line, pos.Column)
	} else if structType, ok = named.Underlying().(*types.Struct); !ok {
		return nil, fmt.Errorf("expected struct type but found %T on field at %s:%d:%d", named.Underlying(), pos.Filename, pos.Line, pos.Column)
	} else if ts, _, ok = r.getType(named.Obj().Name(), named.Obj().Pkg().Path()); !ok {
		panic(fmt.Sprintf("did't find typespec for field defined at %s:%d:%d", pos.Filename, pos.Line, pos.Column))
	} else if structNode, ok = ts.GetTypeSpec().Type.(*dst.StructType); !ok {
		return nil, fmt.Errorf("expected ast node to be StructType but it was %T at %s:%d:%d", ts.GetTypeSpec().Type, pos.Filename, pos.Line, pos.Column)
	} else {
		return r.visitStructFields(parent, ts.Pkg(), structType, structNode)
	}
}

func (r *Registry) visitStructFields(parent *Node, pkg *decorator.Package, structTyp *types.Struct, structDecl *dst.StructType) (nodes []*Node, err error) {
	for i := 0; i < structTyp.NumFields(); i++ {
		var (
			field     = structTyp.Field(i)
			fieldDecl = structDecl.Fields.List[i]
			typ       types.Type
			expr      dst.Expr
			fpkg      *decorator.Package
		)
		if field.Embedded() {
			var tempNodes []*Node
			if tempNodes, err = r.visitEmbeddedStructField(parent, field); err != nil {
				return nil, err
			} else {
				nodes = append(nodes, tempNodes...)
			}
			continue
		}
		typ, expr, fpkg, err = r.dereferenceAndIdentify(field.Type(), fieldDecl.Type, pkg)
		var (
			newNode = &Node{
				Type: typ,
				Node: expr,
				Pkg:  fpkg,
			}
			ts TypeSpec
		)
		if newNode.structField, err = parseFieldConf(field, fieldDecl, parent.Pkg); err != nil {
			return nil, err
		}
		if newNode.structField.ignore() {
			continue
		}
		if fType, ok := field.Type().(*types.Named); ok {
			ts, _, _ = r.getType(fType.Obj().Name(), fType.Obj().Pkg().Path())
			newNode.Pkg = ts.Pkg()
		}
		if newNode.ID, err = r.resolveType(field.Type(), fieldDecl.Type, parent.Pkg); err != nil {
			return nil, err
		}
		//inspect("Adding a struct field: ", field.Name(), field.Type())

		nodes = append(nodes, newNode)
	}
	return nodes, nil
}

func (r *Registry) visitStructTypeNode(parent *Node) ([]*Node, error) {
	var (
		structTyp  = assertType[*types.Struct](parent.Type)
		structDecl = assertType[*dst.StructType](parent.Node)
	)
	return r.visitStructFields(parent, parent.Pkg, structTyp, structDecl)
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
func (r *Registry) visitNamedTypeNode(parent *Node) ([]*Node, error) {
	var (
		node         = &Node{}
		namedType    = parent.Type.(*types.Named)
		typeSpecNode *dst.TypeSpec
		underlying   types.Type
		tsExpr       dst.Expr
		err          error
	)

	switch _tsNode := parent.Node.(type) {
	case *dst.TypeSpec:
		typeSpecNode = _tsNode
	case *dst.Ident:
		if ts, err := r.resolveTypeIdent(_tsNode, parent.Pkg); err != nil {
			return nil, err
		} else if ts == nil {
			panic("oh no: " + _tsNode.String() + " " + parent.Pkg.PkgPath)
		} else if ts.GetType().String() != namedType.String() {
			panic("all bad")
		} else {
			typeSpecNode = ts.GetTypeSpec()
		}
	default:
		return nil, fmt.Errorf("expect types.Name to pair with *dst.TypeSpec. Got %v %T (%v): %w", namedType.Obj(), parent.Node, parent.Node, ErrUnsupportedType)
	}

	if node.ID, err = r.resolveType(namedType.Underlying(), typeSpecNode.Type, parent.Pkg); err != nil {
		return nil, err
	}

	if underlying, tsExpr, node.Pkg, err = r.dereferenceAndIdentify(namedType.Underlying(), typeSpecNode.Type, parent.Pkg); err != nil {
		return nil, err
	}

	switch t := tsExpr.(type) {
	case *dst.Ident:
		switch underlying := underlying.(type) {
		case *types.Basic, *types.Named:
			node.Type = underlying
			node.Node = t
		}
	case *dst.StructType:
		node.Type = underlying
		node.Node = t
	case *dst.ArrayType:
		switch underlying := underlying.(type) {
		case *types.Slice, *types.Array:
			node.Type = underlying
		default:
			panic("should have gotten an array type here.")
		}
		node.Node = t
	case *dst.StarExpr:
		panic("Star Expr should not be possible here.")
	default:
		panic(fmt.Sprintf("unexpected TypeSpec.Expr (%T) %v", t, t))
	}

	switch underlying.(type) {
	case *types.Pointer:
		panic("shouldn't get a pointer here.")
	}
	return []*Node{node}, nil
}

// If pointer type, dereference.
// If Node is an *dst.Ident, locate the type declaration.
func (r *Registry) dereferenceAndIdentify(typ types.Type, expr dst.Expr, localPkg *decorator.Package) (types.Type, dst.Expr, *decorator.Package, error) {
	switch _typ := typ.(type) {
	case *types.Basic:
		return typ, assertType[*dst.Ident](expr), localPkg, nil
	case *types.Named:
		return typ, assertType[*dst.Ident](expr), localPkg, nil
	case *types.Pointer:
		return r.dereferenceAndIdentify(_typ.Elem(), assertType[*dst.StarExpr](expr).X, localPkg)
	}

	// If Type is not a *types.Basic and Node is an ident, I think that means
	// Type should be a named type definition.
	// If Node is not ident, maybe we'll do a little pattern matching
	// just to assert types match.
	switch exprType := expr.(type) {
	case *dst.StarExpr:
		panic("got a star Node after already checking for pointer type")
	case *dst.Ident:
		ts, err := r.locateType(exprType, localPkg)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to dereference type: %w", err)
		}
		return typ, ts.GetTypeSpec().Type, ts.Pkg(), nil
	}

	return typ, expr, localPkg, nil
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

type StructFieldConf struct {
	JSONTagOptions []string
	FieldName      string
	*dst.Field
	*types.Var
	Pkg *decorator.Package
}

func (f StructFieldConf) ignore() bool {
	return f.FieldName == "-" || (f.Tag == nil && !f.Var.Exported())
}

// TypeExpr is the expression node representing the node.
// It should match FieldType, e.g.:
// FieldType = *types.Struct, TypeExpr = *dst. StructType
func (f StructFieldConf) TypeExpr() dst.Expr {
	return f.Field.Type
}

// FieldType is the type object for this fields's type.
func (f StructFieldConf) FieldType() types.Type {
	return f.Var.Type()
}

func (f StructFieldConf) Position() token.Position {
	return f.Pkg.Fset.Position(f.Pos())
}

func (f StructFieldConf) PositionString() string {
	p := f.Position()
	return fmt.Sprintf("at %s:%d:%d", p.Filename, p.Line, p.Column)
}

func parseFieldConf(field *types.Var, fieldNode *dst.Field, pkg *decorator.Package) (*StructFieldConf, error) {
	var (
		fieldName      string
		jsonTagOptions []string
	)
	if !field.Embedded() && fieldNode.Tag != nil {
		tagValue := strings.Trim(fieldNode.Tag.Value, "`")
		if tags, err := structtag.Parse(tagValue); err != nil {
			position := pkg.Fset.Position(field.Pos())

			return &StructFieldConf{}, fmt.Errorf(
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
	return &StructFieldConf{
		JSONTagOptions: jsonTagOptions,
		FieldName:      fieldName,
		Var:            field,
		Field:          fieldNode,
		Pkg:            pkg,
	}, nil
}
