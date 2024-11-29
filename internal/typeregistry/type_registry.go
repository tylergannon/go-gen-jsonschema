package typeregistry

import (
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"log"
)

type Registry struct {
	packages   []*decorator.Package
	typeMap    map[string]TypeSpec
	unionTypes []*UnionTypeDecl
}

type TypeSpec interface {
	TypeSpec() *dst.TypeSpec
	Pkg() *decorator.Package
	GenDecl() *dst.GenDecl
	File() *dst.File
	ID() string
}

type TypeAlternative struct {
	Alias    string
	PkgPath  string
	TypeName string
	FuncName string
}

type UnionTypeDecl struct {
	importMap           ImportMap
	DestTypePackagePath string
	DestTypeName        string
	Alternatives        []TypeAlternative
	File                *dst.File
	Pkg                 *decorator.Package
}

func NewUnionTypeDecl(importMap ImportMap, expr dst.Expr) *UnionTypeDecl {
	switch expr := expr.(type) {
	case *dst.Ident:
		//log.Printf("Name: %s, Path: %s, Obj: %v, %T", expr.Name, expr.Path, expr.Obj, expr.Obj)
		return &UnionTypeDecl{
			importMap:           importMap,
			DestTypePackagePath: importMap[""],
			DestTypeName:        expr.Name,
		}
	default:
		log.Printf("Expr: %T, %v", expr, expr)
	}
	return &UnionTypeDecl{
		importMap: importMap,
	}
}

type typeSpec struct {
	// The type spec for the indicated type
	typeSpec *dst.TypeSpec
	// The package containing the indicated type
	pkg     *decorator.Package
	genDecl *dst.GenDecl
	file    *dst.File
}

func (ts *typeSpec) ID() string {
	return fmt.Sprintf("%s.%s", ts.pkg.PkgPath, ts.typeSpec.Name.Name)
}

var _ TypeSpec = (*typeSpec)(nil)

func (ts *typeSpec) TypeSpec() *dst.TypeSpec {
	return ts.typeSpec
}

func (ts *typeSpec) Pkg() *decorator.Package {
	return ts.pkg
}

func (ts *typeSpec) GenDecl() *dst.GenDecl {
	return ts.genDecl
}

func (ts *typeSpec) File() *dst.File {
	return ts.file
}
