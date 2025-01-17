package syntax

import (
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/tylergannon/structtag"
	"go/token"
	"slices"
	"strings"
	"unicode"
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
		Details() string
	}

	STExpr[T dst.Expr] struct {
		STNode[T]
	}

	Expr interface {
		Node
		NewExpr(expr dst.Expr) Expr
		Expr() dst.Expr
	}

	// TypeExpr is an Expr that is specifically an except from
	// a type declaration.
	// The main idea is mostly to anchor the expr to a specific TypeID.
	// Not 100% sure that's necessary, strictly speaking, but the goal here
	// is more to ensure that there's sufficient information to build
	// the schemas and their associated artifacts.
	TypeExpr struct {
		*TypeSpec
		Excerpt dst.Expr
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
		TypeExpr
		Expr *dst.StructType
	}

	StructField struct {
		STNode[*dst.Field]
	}

	VarConstDecl struct {
		STNode[*dst.GenDecl]
	}
	FuncDecl struct {
		STNode[*dst.FuncDecl]
	}
)

func (s STNode[T]) Details() string {
	return fmt.Sprintf("%T %v", s.node, s.node)
}

func (s STExpr[T]) NewExpr(expr dst.Expr) Expr {
	return NewExpr(expr, s.pkg, s.file)
}

func buildComments(node dst.Node, genDecl *dst.GenDecl) string {
	var comments = BuildComments(node.Decorations())
	if len(comments) > 0 {
		return comments
	}
	if len(genDecl.Specs) > 1 {
		return ""
	}
	return BuildComments(genDecl.Decorations())
}

func NewFuncDecl(f *dst.FuncDecl, pkg *decorator.Package, file *dst.File) FuncDecl {
	return FuncDecl{
		NewNode(f, pkg, file),
	}
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

func NewCallExpr(ce *dst.CallExpr, pkg *decorator.Package, file *dst.File) CallExpr {
	return CallExpr{
		NewExpr(ce, pkg, file),
	}
}

func (c CallExpr) Args() []Expr {
	var args = make([]Expr, len(c.node.Args))
	for i, arg := range c.node.Args {
		args[i] = NewExpr(arg, c.pkg, c.file)
	}
	return args
}

func (e CallExpr) MustIdentifyFunc() TypeID {
	if t, ok := e.IdentifyFunc(); !ok {
		panic("expected type identifier for call expression")
	} else {
		return t
	}
}

func (e CallExpr) IdentifyFunc() (typeID TypeID, ok bool) {
	var expr dst.Expr
	switch _expr := e.node.Fun.(type) {
	case *dst.IndexExpr:
		expr = _expr.X
	default:
		expr = _expr
	}
	switch t := expr.(type) {
	case *dst.SelectorExpr:
		var xIdent *dst.Ident
		if xIdent, ok = t.X.(*dst.Ident); !ok {
			return typeID, false
		}
		typeID.PkgPath, _ = e.Imports().GetPackageForPrefix(xIdent.Name)
		typeID.TypeName = t.Sel.Name
		return typeID, true
	case *dst.Ident:
		if t.Path == "" {
			typeID.PkgPath = e.Pkg().PkgPath
		} else {
			typeID.PkgPath = t.Path
		}
		typeID.TypeName = t.Name
		return typeID, true
	default:
		return TypeID{}, false
	}
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

func (t TypeSpec) Name() string {
	return t.node.Name.Name
}

func (t TypeSpec) Derive() TypeExpr {
	return TypeExpr{TypeSpec: &t, Excerpt: t.node.Type}
}

func (t TypeSpec) Comments() string {
	return buildComments(t.node, t.GenDecl.node)
}

func (t TypeSpec) ID() TypeID {
	return TypeID{PkgPath: t.pkg.PkgPath, TypeName: t.node.Name.Name}
}

/**
 * TypeExpr methods
 */
func (t TypeExpr) Derive(expr dst.Expr) TypeExpr {
	return TypeExpr{TypeSpec: t.TypeSpec, Excerpt: expr}
}

func (t TypeExpr) Pos() token.Pos {
	return t.ToExpr().Pos()
}

func (t TypeExpr) ToExpr() Expr {
	return NewExpr(t.Excerpt, t.pkg, t.file)
}

func (t TypeExpr) Position() token.Position {
	return t.pkg.Fset.Position(t.Pos())
}

/**
 * StructType methods
 */

var NoStructType = StructType{}

func NewStructType(s *dst.StructType, t TypeExpr) StructType {
	return StructType{
		TypeExpr: t,
		Expr:     s,
	}
}

func (s StructType) Fields() (fields []StructField) {
	for _, _field := range s.Expr.Fields.List {
		field := StructField{NewNode(_field, s.pkg, s.file)}
		if field.Skip() {
			continue
		}
		fields = append(fields, field)
	}
	return fields
}

/**
 * StructField methods
 */

func (f StructField) Comments() string {
	return BuildComments(f.node.Decorations())
}

func (f StructField) Embedded() bool {
	return len(f.node.Names) == 0
}

func (f StructField) Type() dst.Expr {
	return f.node.Type
}

func (f StructField) PropNames() (names []string) {
	switch len(f.node.Names) {
	case 0:
		return
	case 1:
		if tag := f.JSONTag(); tag != nil {
			return []string{tag.Options[0]}
		}
	}
	for _, n := range f.node.Names {
		if unicode.IsUpper(rune(n.Name[0])) {
			names = append(names, n.Name)
		}
	}
	return names
}

func (f StructField) JSONTag() *structtag.Tag {
	if f.node.Tag == nil {
		return nil
	} else if tags, err := structtag.Parse(strings.Trim(f.node.Tag.Value, "`")); err == nil {
		if tag, err := tags.Get("json"); err == nil && len(tag.Options) > 0 {
			return tag
		}
	}
	return nil
}

func (f StructField) Skip() bool {
	if len(f.node.Names) == 0 {
		return false
	} else if !slices.ContainsFunc(f.node.Names, func(ident *dst.Ident) bool {
		return unicode.IsUpper(rune(ident.Name[0]))
	}) {
		return true
	} else if tag := f.JSONTag(); tag != nil {
		return tag.Options[0] == "-"
	}
	return false
}

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

/**
 * NamedType methods
 */

func (n TypeSpec) Type() Expr {
	return NewExpr(n.node.Type, n.pkg, n.file)
}
