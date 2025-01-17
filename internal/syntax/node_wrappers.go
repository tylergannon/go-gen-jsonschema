package syntax

import (
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"go/token"
)

type (
	STNode[T dst.Node] struct {
		node T
		file *dst.File
		pkg  *decorator.Package
	}

	Node interface {
		Node() dst.Node
		File() *dst.File
		Pkg() *decorator.Package
		Pos() token.Pos
		Position() token.Position
		Imports() ImportMap
	}

	STExpr[T dst.Expr] struct {
		STNode[T]
	}

	Expr interface {
		Node
		Expr() dst.Expr
	}

	CallExpr struct {
		STExpr[*dst.CallExpr]
	}

	ValueSpec struct {
		GenDecl STNode[*dst.GenDecl]
		STNode[*dst.ValueSpec]
	}

	TypeSpec struct {
		GenDecl STNode[*dst.GenDecl]
		STNode[*dst.TypeSpec]
	}

	StructType struct {
		STNode[*dst.StructType]
	}

	VarConstDecl struct {
		STNode[*dst.GenDecl]
	}
)

func buildComments(node dst.Node, genDecl *dst.GenDecl) string {
	var comments = BuildComments(node.Decorations())
	if len(comments) > 0 {
		return comments
	}
	if len(genDecl.Specs) > 0 {
		return ""
	}
	return BuildComments(genDecl.Decorations())
}

func NewVarConstDecl(node *dst.GenDecl, pkg *decorator.Package, file *dst.File) VarConstDecl {
	return VarConstDecl{
		STNode[*dst.GenDecl]{
			node: node,
			file: file,
			pkg:  pkg,
		},
	}
}

/**
 * STNode methods
 */

func (s STNode[T]) Node() dst.Node {
	return s.node
}

func (s STNode[T]) File() *dst.File {
	return s.file
}

func (s STNode[T]) Pkg() *decorator.Package {
	return s.pkg
}

func (s STNode[T]) Pos() token.Pos {
	return s.pkg.Decorator.Map.Ast.Nodes[s.node].Pos()
}

func (s STNode[T]) Position() token.Position {
	return s.pkg.Fset.Position(s.Pos())
}

func (s STNode[T]) Imports() ImportMap {
	return s.file.Imports
}

func NewNode[T dst.Node](node T, pkg *decorator.Package, file *dst.File) STNode[T] {
	return STNode[T]{
		node: node,
		file: file,
		pkg:  pkg,
	}
}

var _ Node = STNode[dst.Node]{}

/**
 * STExpr methods
 */

func (s STExpr[T]) Expr() dst.Expr {
	return s.node
}

func NewExpr[T dst.Expr](expr T, pkg *decorator.Package, file *dst.File) STExpr[T] {
	return STExpr[T]{
		STNode: NewNode(expr, pkg, file),
	}
}

var _ Expr = STExpr[dst.Expr]{}

/**
 * CallExpr methods
 */

func (c CallExpr) Args() []Expr {
	var args = make([]Expr, len(c.node.Args))
	for i, arg := range c.node.Args {
		args[i] = NewExpr(arg, c.pkg, c.file)
	}
	return args
}

/**
 * ValueSpec methods
 */

func NewValueSpec(genDecl *dst.GenDecl, node *dst.ValueSpec, pkg *decorator.Package, file *dst.File) ValueSpec {
	return ValueSpec{
		GenDecl: NewNode(genDecl, pkg, file),
		STNode:  NewNode(node, pkg, file),
	}
}

func (v ValueSpec) Comments() string {
	return buildComments(v.node, v.GenDecl.node)
}

func (v ValueSpec) HasType() bool {
	return v.node.Type != nil
}

func (t ValueSpec) Type() dst.Expr {
	return t.node.Type
}

func (t ValueSpec) Value() *dst.ValueSpec {
	return t.node
}

/**
 * TypeSpec methods
 */

func NewTypeSpec(genDecl *dst.GenDecl, ts *dst.TypeSpec, pkg *decorator.Package, file *dst.File) TypeSpec {
	return TypeSpec{
		GenDecl: NewNode(genDecl, pkg, file),
		STNode:  NewNode(ts, pkg, file),
	}
}

func (t TypeSpec) Comments() string {
	return buildComments(t.node, t.GenDecl.node)
}

func (t TypeSpec) ID() TypeID {
	return TypeID{PkgPath: t.pkg.PkgPath, TypeName: t.node.Name.Name}
}

/**
 * StructType methods
 */

/**
 * VarConstDecl methods
 */

func (v VarConstDecl) Specs() []ValueSpec {
	var specs = make([]ValueSpec, len(v.node.Specs))
	for i, spec := range v.node.Specs {
		specs[i] = ValueSpec{
			GenDecl: v.STNode,
			STNode:  NewNode(spec.(*dst.ValueSpec), v.pkg, v.file),
		}
	}
	return specs
}
