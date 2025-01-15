package scanner

import (
	"fmt"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
)

type PackageScanner interface {
	Scan(*packages.Package) (ScanResult, error)
}

// ScanResult aggregates the important details from a single package:
// 1) All the (named, as in GenDecl) type declarations
// 2) All the const decls (filtered by which ones are marked)
// 3) All the interface implementations (filtered by which ones are marked)
// 4)
type ScanResult struct {
	Pkg       *packages.Package
	TypeDecls []TypeDecl
	Constants []constDecl
}

// TypeDecl refers to the
type TypeDecl struct {
	Node  *ast.DeclStmt
	Pkg   *packages.Package
	File  *ast.File
	Pos   token.Pos
	Decls []TypeInfo
}

type InterfaceTypeInfo struct {
	typeInfo
}

// AliasTypeInfo
// refers to type alias declarations.
type AliasTypeInfo struct {
	typeInfo
}

// DefinitionTypeInfo
// A Definition type Refers to any type that's declared as an instance of another type,
// including type-instantiated generics.
// e.g.:
// ```
// type MyInt int
// type MyString string
// type MyFooType somepackage.SomeGenericType[MyInt, MyType]
// ```
type DefinitionTypeInfo struct {
	typeInfo
}

var _ TypeInfo = InterfaceTypeInfo{}

type typeInfo struct {
	Identity *ast.Ident
	TypeInfo ast.Expr
	File     *ast.File
	FileSet  *token.FileSet
	Pkg      *packages.Package
	Decl     *TypeDecl
}

type TypeInfo interface {
	GetTypeName() string
	GetPkgPath() string
	GetTypeInfo() ast.Expr
	GetFile() *ast.File
	GetPkg() *packages.Package
	GetDecl() *TypeDecl
}

func (t typeInfo) GetDecl() *TypeDecl {
	return t.Decl
}

// Getters for typeInfo fields.
func (t typeInfo) GetTypeName() string {
	return t.Identity.Name
}

func (t typeInfo) GetPkgPath() string {
	return t.Pkg.PkgPath
}

func (t typeInfo) GetTypeInfo() ast.Expr {
	return t.TypeInfo
}

func (t typeInfo) GetFile() *ast.File {
	return t.File
}

func (t typeInfo) GetPkg() *packages.Package {
	return t.Pkg
}

const (
	normalFunc FuncType = iota
	concreteReceiver
	pointerReceiver
)

type (
	TypeDecls struct {
		Pkg  *packages.Package
		File *ast.File

		Decl  *ast.GenDecl
		Specs []*ast.TypeSpec
	}

	VarDecls struct {
		Pkg   *packages.Package
		File  *ast.File
		Decl  *ast.GenDecl
		Specs []*ast.ValueSpec
	}

	FuncDecl struct {
		Pkg  *packages.Package
		File *ast.File
		Decl *ast.FuncDecl
	}

	ConstDecls struct {
		Pkg   *packages.Package
		File  *ast.File
		Decl  *ast.GenDecl
		Specs []*ast.ValueSpec
	}

	FuncType       int
	schemaFuncInfo struct {
		FuncType FuncType
		FuncName string
	}
	ifaceImplInfo struct {
		// PkgPath is empty string to denote the local package
		PkgPath  string
		TypeName string
		Pointer  bool
	}
	// ifaceImplementations represents an interface type and its allowed types.
	// Error in the event that the loader encounters TypeName referenced on any
	// types declared outside of this package.
	// That is to say, in order for an interface implementations to work,
	// all supported references to it must be in the local package.
	ifaceImplementations struct {
		PkgPath  string
		TypeName string
		Impls    []ifaceImplInfo
	}

	constDecl struct {
		TypeName string
		Values   []*ast.BasicLit
	}
)

func (s schemaFuncInfo) markerType() MarkerKind {
	return MarkerKindSchema
}

func (c constDecl) markerType() MarkerKind {
	return MarkerKindEnum
}

func (i ifaceImplementations) markerType() MarkerKind {
	return MarkerKindInterface
}

// Marker is the common interface for all returned marker structs.
type Marker interface {
	markerType() MarkerKind
}

var (
	_ Marker = ifaceImplementations{}
	_ Marker = schemaFuncInfo{}
	_ Marker = constDecl{}
)

// MarkerKind enumerates the four categories of markers we have.
type MarkerKind string

const (
	MarkerKindEnum      MarkerKind = "Enum"
	MarkerKindInterface MarkerKind = "InterfaceImpl"
	MarkerKindSchema    MarkerKind = "Schema"
)

func DecodeFuncCall(callExpr *ast.CallExpr) (Marker, bool) {

	switch expr := callExpr.Fun.(type) {
	case *ast.SelectorExpr:
		fmt.Printf("SelectorExpr -- %T, %v, %v\n", expr.X, expr.X, expr.Sel)
	case *ast.IndexExpr:
		fmt.Printf("IndexExpr -- %T, %v, %v\n", expr.X, expr.X, expr.Index)
	}
	return nil, false
}

func LoadPackage(pkg *packages.Package) {
	var (
		constDecls []ConstDecls
		typeDecls  []TypeDecls
		varDecls   []VarDecls
		funcDecls  []FuncDecl
	)
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			switch _decl := decl.(type) {
			case *ast.FuncDecl:
				funcDecls = append(funcDecls, FuncDecl{
					Pkg:  pkg,
					File: file,
					Decl: _decl,
				})
			case *ast.GenDecl:
				switch _decl.Tok {
				case token.TYPE:
					var specs []*ast.TypeSpec
					for _, spec := range _decl.Specs {
						specs = append(specs, spec.(*ast.TypeSpec))
					}
					typeDecls = append(typeDecls, TypeDecls{
						Pkg:   pkg,
						File:  file,
						Decl:  _decl,
						Specs: specs,
					})
				case token.CONST:
					var values []*ast.ValueSpec
					for _, spec := range _decl.Specs {
						values = append(values, spec.(*ast.ValueSpec))
					}
					constDecls = append(constDecls, ConstDecls{
						Pkg:   pkg,
						File:  file,
						Decl:  _decl,
						Specs: values,
					})
				case token.VAR:
					var specs []*ast.ValueSpec
					for _, spec := range _decl.Specs {
						specs = append(specs, spec.(*ast.ValueSpec))
					}
					varDecls = append(varDecls, VarDecls{
						Pkg:   pkg,
						File:  file,
						Decl:  _decl,
						Specs: specs,
					})
				default:
				}
			}
		}
	}
}
