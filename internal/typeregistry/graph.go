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

type Node interface {
	Type() types.Type
	DSTNode() dst.Node
	Pkg() *decorator.Package
	ID() TypeID
	// Inbound is the count of edges into this node
	Inbound() int
}

type nodeInternal interface {
	Node
	setInbound(int)
	addChild(id TypeID)
	Children() []TypeID
}

type NamedTypeNode struct {
	*nodeImpl
	IsAlt    bool // Whether this particular type instance is intended as an alternate in a Union type
	TypeSpec TypeSpec
}

func (n NamedTypeNode) Name() string {
	return n.NamedType().Obj().Name()
}

func (n NamedTypeNode) UnderlyingTypeID() TypeID {
	if len(n.children) != 1 {
		panic(fmt.Sprintf("unexpected number of children: %d", len(n.children)))
	}
	return n.children[0]
}

func (n NamedTypeNode) NamedType() *types.Named {
	return n.typ.(*types.Named)
}

func (n NamedTypeNode) TypeSpecNode() *dst.TypeSpec {
	return n.dstNode.(*dst.TypeSpec)
}

type InterfaceTypeNode struct {
	NamedTypeNode
	Implementations []InterfaceImpl
}

// NamedTypeWithAltsNode represents a NamedTypeNode that's specified according
// to its alternative types.
// The alternatives themselves are defined on the TypeSpec.Alternatives().
type NamedTypeWithAltsNode struct {
	NamedTypeNode
}

type StructTypeNode struct {
	*nodeImpl
}

func (n StructTypeNode) Fields() []TypeID {
	return n.children
}

type SliceTypeNode struct {
	*nodeImpl
}

type StructFieldNode struct {
	fieldType   nodeInternal
	FieldConf   *StructFieldConf
	parentID    TypeID
	pkg         *decorator.Package
	FieldTypeID TypeID
}

func (n StructFieldNode) Type() types.Type {
	//TODO implement me
	panic("implement me")
}

func (n StructFieldNode) DSTNode() dst.Node {
	return n.FieldConf.Field
}

func (n StructFieldNode) Pkg() *decorator.Package {
	return n.pkg
}

func (n StructFieldNode) Inbound() int {
	return 1
}

func (n StructFieldNode) setInbound(int) {
}

func (n StructFieldNode) addChild(TypeID) {
}

func (n StructFieldNode) Children() []TypeID {
	return []TypeID{n.FieldTypeID}
}

var _ nodeInternal = StructFieldNode{}

func (n StructFieldNode) ID() TypeID {
	return TypeID(fmt.Sprintf("%s!%s", n.parentID, n.FieldConf.FieldName))
}

func (n SliceTypeNode) ElemNodeID() TypeID {
	if len(n.children) != 1 {
		panic(fmt.Sprintf("unexpected number of children: %d", len(n.children)))
	}
	return n.children[0]
}

type EnumTypeNode struct {
	NamedTypeNode
	Entries []EnumEntry
}

type BasicTypeNode struct {
	*nodeImpl
}

func (n BasicTypeNode) BasicType() *types.Basic {
	return n.typ.(*types.Basic)
}

var _ nodeInternal = NamedTypeNode{}

type nodeImpl struct {
	//parent   *nodeImpl              // the nodeImpl that caused this to go onto the stack
	typ      types.Type         // the types object
	dstNode  dst.Node           // should be either a TypeSpec or a Expr, depending on whether the Type is a named type or not.
	pkg      *decorator.Package // the decorated package that this nodeImpl occurs in
	children []TypeID           // before leaving from the first visit, must be initialized to have length equal to number of child nodes
	inbound  int                // Number of edges with this node as destination
	// The following are determined on the second visit.
	id TypeID // update this after resolving the type
}

func (n *nodeImpl) addChild(id TypeID) {
	n.children = append(n.children, id)
}

func (n *nodeImpl) setInbound(inbound int) {
	n.inbound = inbound
}

func (n *nodeImpl) Children() []TypeID {
	return n.children
}

func (n *nodeImpl) Type() types.Type {
	return n.typ
}

func (n *nodeImpl) DSTNode() dst.Node {
	return n.dstNode
}

func (n *nodeImpl) Pkg() *decorator.Package {
	return n.pkg
}

func (n *nodeImpl) ID() TypeID {
	return n.id
}

func (n *nodeImpl) Inbound() int {
	return n.inbound
}

func countInbound(nodes map[TypeID]nodeInternal) map[TypeID]Node {
	var result = map[TypeID]Node{}
	counts := make(map[TypeID]int, len(nodes))
	for _, niceGuy := range nodes {
		for _, child := range niceGuy.Children() {
			counts[child]++
		}
	}
	for id, n := range nodes {
		n.setInbound(counts[id])
		result[id] = n
	}
	return result
}

type SchemaGraph struct {
	RootNode Node
	Nodes    map[TypeID]Node
}

func (r *Registry) GraphTypeForSchema(ts TypeSpec) (*SchemaGraph, error) {
	var (
		nodes    = map[TypeID]nodeInternal{}
		rootNode = r.buildNodeForTS(ts, false)
		stack    = []nodeInternal{rootNode}
		parents  = map[TypeID]nodeInternal{}
	)

	for i := 0; len(stack) > 0; i++ {
		if i > 10000 {
			return nil, fmt.Errorf("iterated a little long, perhaps")
		}
		node := stack[0]
		stack = stack[1:]
		if _, ok := nodes[node.ID()]; ok {
			continue
		}
		nodes[node.ID()] = node

		if nodesTemp, err := r.visitNode(node); err != nil {
			return nil, err
		} else {
			if iface, ok := node.(*InterfaceTypeNode); ok {
				// assert parent node must be a StructField, and its grand. parent
				// must be a named type.
				// And *that* named type must be given an unmarshaler.  Meaning,
				// the type must be in the local package or else it must already
				// have an unmarshaler function.
				_ = iface
			}
			for _, nodeTemp := range nodesTemp {
				parents[nodeTemp.ID()] = node
				node.addChild(nodeTemp.ID())
				if _, ok := nodes[nodeTemp.ID()]; !ok {
					stack = append(stack, nodeTemp)
				}
			}
		}
	}
	return &SchemaGraph{RootNode: rootNode, Nodes: countInbound(nodes)}, nil
}

func (r *Registry) resolveTypeIdent(ident *dst.Ident, currPkgPath *decorator.Package) (TypeSpec, error) {
	pkgPath := ident.Path
	if ident.Path == "" {
		pkgPath = currPkgPath.PkgPath
	}
	ts, ok := r.getType(ident.Name, pkgPath)
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
func (r *Registry) visitNode(node nodeInternal) ([]nodeInternal, error) {
	switch tn := node.(type) {
	case NamedTypeNode:
		return r.visitNamedTypeNode(tn)
	case NamedTypeWithAltsNode:
		return r.visitTypeWithAltsNode(tn)
	case InterfaceTypeNode:
		return r.visitInterfaceTypeNode(tn)
	case StructTypeNode:
		var result []nodeInternal
		if _temp, _err := r.visitStructTypeNode(tn); _err != nil {
			return nil, _err
		} else {
			for _, temp := range _temp {
				result = append(result, temp)
			}
		}
		return result, nil
	case SliceTypeNode:
		return r.visitSliceTypeNode(tn)
	case StructFieldNode:
		return []nodeInternal{tn.fieldType}, nil
	case BasicTypeNode, EnumTypeNode:
		return nil, nil
	}

	switch node.Type().(type) {
	case *types.Alias:
		return nil, fmt.Errorf("alias types not yet supported: %w", ErrUnsupportedType)
	case *types.Pointer:
		fmt.Println(string(debug.Stack()))
		panic("we don't do pointers here")
	case *types.Basic:
		return nil, nil
	default:
		inspect("unexpected node", node.DSTNode(), node.Type(), node.ID(), node)

		panic(fmt.Sprintf("unexpected node type: %T", node.Type()))
	}
}

func (r *Registry) visitTypeWithAltsNode(node NamedTypeWithAltsNode) ([]nodeInternal, error) {
	var result []nodeInternal
	alts := node.TypeSpec.Alternatives()
	if len(alts) == 1 {
		result = append(result, r.buildNodeForTS(alts[0].TypeSpec, false))
		return result, nil
	}
	for _, alt := range node.TypeSpec.Alternatives() {
		result = append(result, r.buildNodeForTS(alt.TypeSpec, true))
	}

	return result, nil
}

func (r *Registry) visitInterfaceTypeNode(node InterfaceTypeNode) ([]nodeInternal, error) {
	var result []nodeInternal

	for _, alt := range node.Implementations {
		if ts, ok := r.getType(alt.TypeName, alt.PkgPath); !ok {
			return nil, fmt.Errorf("type identifier not found: %s", alt.TypeName)
		} else {
			result = append(result, r.buildNodeForTS(ts, true))
		}
	}

	return result, nil
}

func (r *Registry) visitSliceTypeNode(node SliceTypeNode) ([]nodeInternal, error) {
	var (
		t           = node.Type().(*types.Slice)
		ts          = node.DSTNode().(*dst.ArrayType)
		typeID, err = r.resolveType(t.Elem(), ts.Elt, node.Pkg())
	)
	if err != nil {
		return nil, err
	}
	return []nodeInternal{
		r.resolveNodeType(
			&nodeImpl{
				typ:     t.Elem(),
				dstNode: ts.Elt,
				pkg:     node.Pkg(),
				id:      typeID,
			},
		),
	}, nil
}

// visitEmbeddedStructField basically validates the embedded type to be a named type
// parent refers to the node that owns the field itself.
// field is the field var from parent.
func (r *Registry) visitEmbeddedStructField(parent StructTypeNode, field *types.Var) (nodes []StructFieldNode, err error) {
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
	} else if ts, ok = r.getType(named.Obj().Name(), named.Obj().Pkg().Path()); !ok {
		panic(fmt.Sprintf("did't find typespec for field defined at %s:%d:%d", pos.Filename, pos.Line, pos.Column))
	} else if structNode, ok = ts.GetTypeSpec().Type.(*dst.StructType); !ok {
		return nil, fmt.Errorf("expected ast node to be StructType but it was %T at %s:%d:%d", ts.GetTypeSpec().Type, pos.Filename, pos.Line, pos.Column)
	} else {
		return r.visitStructFields(parent, ts.Pkg(), structType, structNode)
	}
}

func (r *Registry) visitStructFields(parent StructTypeNode, pkg *decorator.Package, structTyp *types.Struct, structDecl *dst.StructType) (nodes []StructFieldNode, err error) {
	for i := 0; i < structTyp.NumFields(); i++ {
		var (
			field     = structTyp.Field(i)
			fieldDecl = structDecl.Fields.List[i]
			typ       types.Type
			expr      dst.Expr
			fpkg      *decorator.Package
		)
		if field.Embedded() {
			var tempNodes []StructFieldNode
			if tempNodes, err = r.visitEmbeddedStructField(parent, field); err != nil {
				return nil, err
			} else {
				nodes = append(nodes, tempNodes...)
			}
			continue
		}
		if typ, expr, fpkg, err = r.dereferenceAndIdentify(field.Type(), fieldDecl.Type, pkg); err != nil {
			return nil, err
		}
		var (
			newNode = &nodeImpl{
				typ:     typ,
				dstNode: expr,
				pkg:     fpkg,
			}
			ts          TypeSpec
			structField *StructFieldConf
		)
		if structField, err = parseFieldConf(field, fieldDecl, parent.Pkg()); err != nil {
			return nil, err
		}
		if structField.ignore() {
			continue
		}
		if fType, ok := field.Type().(*types.Named); ok {
			if ts, ok = r.getType(fType.Obj().Name(), fType.Obj().Pkg().Path()); !ok {
				fmt.Printf("Didn't find type %s in package %s\n", fType.Obj().Name(), fType.Obj().Pkg().Path())
			}
			newNode.pkg = ts.Pkg()
		}
		if newNode.id, err = r.resolveType(field.Type(), fieldDecl.Type, parent.Pkg()); err != nil {
			return nil, err
		}

		fieldType := r.resolveNodeType(newNode)
		nodes = append(nodes, StructFieldNode{
			fieldType:   fieldType,
			FieldConf:   structField,
			parentID:    parent.id,
			pkg:         pkg,
			FieldTypeID: fieldType.ID(),
		})
	}
	return nodes, nil
}

func (r *Registry) buildNodeForTS(ts TypeSpec, isAlt bool) nodeInternal {
	var id = ts.ID()
	if isAlt {
		id = TypeID(string(id) + "~")
	}
	node := NamedTypeNode{
		TypeSpec: ts,
		IsAlt:    isAlt,
		nodeImpl: &nodeImpl{
			typ:     ts.GetType(),
			dstNode: ts.GetTypeSpec(),
			pkg:     ts.Pkg(),
			id:      id,
		},
	}

	if iface, ok := r.interfaceTypes[ts.ID()]; ok {
		return InterfaceTypeNode{
			NamedTypeNode:   node,
			Implementations: iface.Implementations,
		}
	}

	if len(ts.Alternatives()) > 0 {
		return NamedTypeWithAltsNode{node}
	}

	return node
}

func (r *Registry) visitStructTypeNode(parent StructTypeNode) ([]StructFieldNode, error) {
	var (
		structTyp  = assertType[*types.Struct](parent.Type())
		structDecl = assertType[*dst.StructType](parent.DSTNode())
	)
	return r.visitStructFields(parent, parent.Pkg(), structTyp, structDecl)
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
func (r *Registry) visitNamedTypeNode(parent NamedTypeNode) ([]nodeInternal, error) {
	var (
		node         = &nodeImpl{}
		namedType    = parent.Type().(*types.Named)
		typeSpecNode *dst.TypeSpec
		underlying   types.Type
		tsExpr       dst.Expr
		err          error
	)

	switch _tsNode := parent.DSTNode().(type) {
	case *dst.TypeSpec:
		typeSpecNode = _tsNode
	case *dst.Ident:
		panic("shouldn't be here.")
		//if ts, err := r.resolveTypeIdent(_tsNode, parent.Pkg()); err != nil {
		//	return nil, err
		//} else if ts == nil {
		//	panic("oh no: " + _tsNode.String() + " " + parent.Pkg().PkgPath)
		//} else if ts.GetType().String() != namedType.String() {
		//	panic("all bad")
		//} else {
		//	typeSpecNode = ts.GetTypeSpec()
		//}
	default:
		return nil, fmt.Errorf(
			"expect types.Name to pair with *dst.TypeSpec. Got %v %T (%v): %w",
			namedType.Obj(),
			parent.DSTNode(),
			parent.DSTNode(),
			ErrUnsupportedType,
		)
	}

	if node.id, err = r.resolveType(namedType.Underlying(), typeSpecNode.Type, parent.Pkg()); err != nil {
		return nil, err
	}

	if underlying, tsExpr, node.pkg, err = r.dereferenceAndIdentify(namedType.Underlying(), typeSpecNode.Type, parent.Pkg()); err != nil {
		return nil, err
	}

	switch t := tsExpr.(type) {
	case *dst.Ident:
		switch underlying := underlying.(type) {
		case *types.Basic, *types.Named:
			node.typ = underlying
			node.dstNode = t
		}
	case *dst.StructType:
		node.typ = underlying
		node.dstNode = t
	case *dst.ArrayType:
		switch underlying := underlying.(type) {
		case *types.Slice, *types.Array:
			node.typ = underlying
		default:
			panic("should have gotten an array type here.")
		}
		node.dstNode = t
	case *dst.StarExpr:
		panic("Star Expr should not be possible here.")
	default:
		panic(fmt.Sprintf("unexpected TypeSpec.Expr (%T) %v", t, t))
	}

	switch underlying.(type) {
	case *types.Pointer:
		panic("shouldn't get a pointer here.")
	}
	return []nodeInternal{r.resolveNodeType(node)}, nil
}

func (r *Registry) resolveNodeType(n *nodeImpl) nodeInternal {
	switch t := n.typ.(type) {
	case *types.Named:
		if tsTemp, ok := r.getType(t.Obj().Name(), t.Obj().Pkg().Path()); !ok {
			panic("why I not find")
		} else {
			n.dstNode = tsTemp.typeSpec
			n.typ = tsTemp.GetType()
			newNode := NamedTypeNode{TypeSpec: tsTemp, nodeImpl: n}
			if entries, ok := r.constants[newNode.id]; ok && len(entries) > 0 {
				return EnumTypeNode{newNode, entries}
			}
			if len(tsTemp.alts) > 0 {
				return NamedTypeWithAltsNode{newNode}
			}
			return newNode
		}
	case *types.Slice, *types.Array:
		return SliceTypeNode{n}
	case *types.Struct:
		return StructTypeNode{n}
	case *types.Basic:
		return BasicTypeNode{n}
	default:
		panic(fmt.Sprintf("unexpected type: %T", t))
	}
}

// If pointer type, dereference.
// If nodeImpl is an *dst.Ident, locate the type declaration.
func (r *Registry) dereferenceAndIdentify(typ types.Type, expr dst.Expr, localPkg *decorator.Package) (types.Type, dst.Expr, *decorator.Package, error) {
	switch _typ := typ.(type) {
	case *types.Basic:
		return typ, assertType[*dst.Ident](expr), localPkg, nil
	case *types.Named:
		return typ, assertType[*dst.Ident](expr), localPkg, nil
	case *types.Pointer:
		return r.dereferenceAndIdentify(_typ.Elem(), assertType[*dst.StarExpr](expr).X, localPkg)
	}

	// If Type is not a *types.Basic and nodeImpl is an ident, I think that means
	// Type should be a named type definition.
	// If nodeImpl is not ident, maybe we'll do a little pattern matching
	// just to assert types match.
	switch exprType := expr.(type) {
	case *dst.StarExpr:
		panic("got a star nodeImpl after already checking for pointer type")
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
	ts, ok := r.getType(ident.Name, path)
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
