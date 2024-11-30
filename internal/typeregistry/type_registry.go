package typeregistry

import (
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"log"
)

type Registry struct {
	packages   map[string]*decorator.Package
	typeMap    map[TypeID]TypeSpec
	unionTypes map[TypeID]*UnionTypeDecl
}

func (r *Registry) AddTarget(typeName, pkgPath string) error {
	if err := r.LoadAndScan(pkgPath); err != nil {
		return fmt.Errorf("loading package: %w", err)
	}
	//ts, unionType, _ := r.GetType(typeName, pkgPath)

	return nil
}

type TypeID string

type IdentifiableByType interface {
	// ID returns a string name composed of package and type name.
	ID() TypeID
}

func NewTypeID(pkgPath, typeName string) TypeID {
	return TypeID(pkgPath + "." + typeName)
}

func (r *Registry) GetType(name string, pkgPath string) (TypeSpec, *UnionTypeDecl, bool) {
	typeID := NewTypeID(pkgPath, name)
	if ts, ok := r.typeMap[typeID]; !ok {
		return ts, r.unionTypes[typeID], true
	}
	return nil, nil, false
}

type TypeSpec interface {
	IdentifiableByType
	TypeSpec() *dst.TypeSpec
	Pkg() *decorator.Package
	GenDecl() *dst.GenDecl
	File() *dst.File
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

func (d *UnionTypeDecl) ID() TypeID {
	return NewTypeID(d.DestTypePackagePath, d.DestTypeName)
}

func SetTypeAlternativeDecl(importMap ImportMap, expr dst.Expr) *UnionTypeDecl {
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

func (ts *typeSpec) ID() TypeID {
	return NewTypeID(ts.pkg.PkgPath, ts.typeSpec.Name.Name)
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
