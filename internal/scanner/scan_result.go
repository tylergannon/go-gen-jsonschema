package scanner

import (
	"fmt"
	"github.com/tylergannon/go-gen-jsonschema/internal/importmap"
	"go/ast"
	"go/token"
	"go/types"
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
	Implementations []TypeID
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

	SchemaMethod struct {
		Receiver   TypeID
		FuncName   string
		MarkerCall MarkerFunctionCall
	}
	SchemaFunction SchemaMethod
	// ifaceImplementations represents an interface type and its allowed types.
	// Error in the event that the loader encounters TypeName referenced on any
	// types declared outside of this package.
	// That is to say, in order for an interface implementations to work,
	// all supported references to it must be in the local package.
	ifaceImplementations struct {
		Pkg    *packages.Package
		File   *ast.File
		TypeID TypeID
		Impls  []TypeID
	}

	enumVal struct {
		GenDecl *ast.GenDecl
		Decl    *ast.ValueSpec
	}

	NamedTypeSpec struct {
		NamedType *types.Named
		TypeSpec  *ast.TypeSpec
		File      *ast.File
		Pkg       *packages.Package
	}

	enumSet struct {
		GenDecl  *ast.GenDecl
		TypeSpec *ast.TypeSpec
		Pkg      *packages.Package
		File     *ast.File
		TypeID   TypeID
		Values   []enumVal
	}
)

func (s SchemaMethod) markerType() MarkerKind {
	return MarkerKindSchema
}

func (c enumSet) markerType() MarkerKind {
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
	_ Marker = SchemaMethod{}
	_ Marker = enumSet{}
)

// MarkerKind enumerates the four categories of markers we have.
type MarkerKind string

const (
	MarkerKindEnum      MarkerKind = "Enum"
	MarkerKindInterface MarkerKind = "InterfaceImpl"
	MarkerKindSchema    MarkerKind = "Schema"
)

func (v VarDecls) importMap() importmap.ImportMap {
	return v.File.Imports
}

type VarDeclSet []VarDecls

func (vd VarDeclSet) MarkerFuncs() []MarkerFunctionCall {
	var result []MarkerFunctionCall
	for _, decl := range vd {
		for _, spec := range decl.Specs {
			result = append(
				result,
				ParseValueExprForMarkerFunctionCall(spec, decl.File, decl.Pkg)...,
			)
		}
	}
	return result
}

type decls struct {
	constDecls []ConstDecls
	typeDecls  []TypeDecls
	varDecls   VarDeclSet
	funcDecls  []FuncDecl
}

type ScanResult struct {
	Pkg             *packages.Package
	Constants       map[string]*enumSet
	MarkerCalls     []MarkerFunctionCall
	Interfaces      map[string]ifaceImplementations
	ConcreteTypes   map[TypeID]bool
	SchemaMethods   []SchemaMethod
	SchemaFuncs     []SchemaFunction
	LocalNamedTypes map[string]NamedTypeSpec
}

func LoadPackage(pkg *packages.Package) (ScanResult, error) {
	// Needs to discover:
	// 1. Enum (Const) Values
	// 2. Supported Interfaces
	// 3. Types to render
	var (
		_decls          = loadPkgDecls(pkg)
		markerCalls     = _decls.varDecls.MarkerFuncs()
		enums           = map[string]*enumSet{}
		interfaces      = map[string]ifaceImplementations{}
		concreteTypes   = map[TypeID]bool{}
		schemaMethods   []SchemaMethod
		schemaFuncs     []SchemaFunction
		localNamedTypes = map[string]NamedTypeSpec{}
	)
	for _, decl := range markerCalls {
		switch decl.Function {
		case MarkerFuncNewEnumType:
			enums[decl.TypeArgument.TypeName] = &enumSet{
				Pkg:    decl.Pkg,
				TypeID: *decl.TypeArgument,
			}
		case MarkerFuncNewInterfaceImpl:
			var (
				err   error
				iface = ifaceImplementations{
					TypeID: *decl.TypeArgument,
				}
			)
			if iface.Impls, err = decl.ParseTypesFromArgs(); err != nil {
				return ScanResult{}, err
			}
			interfaces[iface.TypeID.TypeName] = iface
			for _, impl := range iface.Impls {
				concreteTypes[impl] = true
			}
		case MarkerFuncNewJSONSchemaMethod:
			method, err := decl.ParseSchemaMethod()
			concreteTypes[method.Receiver] = true
			if err != nil {
				return ScanResult{}, err
			}
			schemaMethods = append(schemaMethods, method)
		case MarkerFuncNewJSONSchemaBuilder:
			method, err := decl.ParseSchemaFunc()
			concreteTypes[method.Receiver] = true
			if err != nil {
				return ScanResult{}, err
			}
			schemaFuncs = append(schemaFuncs, method)

		default:
			return ScanResult{}, fmt.Errorf("unsupported marker function: %s", decl.Function)
		}
	}
	for _, _typeDecl := range _decls.typeDecls {
		for _, spec := range _typeDecl.Specs {
			var (
				typeID = TypeID{DeclaredLocally: true, TypeName: spec.Name.Name}
			)
			if iface, ok := interfaces[typeID.TypeName]; ok {
				iface.Pkg = pkg
				iface.File = _typeDecl.File
				interfaces[typeID.TypeName] = iface
			} else if enum, ok := enums[typeID.TypeName]; ok {
				enum.GenDecl = _typeDecl.Decl
				enum.TypeSpec = spec
				enum.File = _typeDecl.File
			} else {
				var t = NamedTypeSpec{
					TypeSpec: spec,
					File:     _typeDecl.File,
					Pkg:      _typeDecl.Pkg,
				}
				if t.NamedType, ok = findNamedType(_typeDecl.Pkg, spec.Name.Name); !ok {
					pos := _typeDecl.Pkg.Fset.Position(spec.Pos())
					return ScanResult{}, fmt.Errorf("unable to load named type for %s declared at %s", spec.Name.Name, pos)
				}
				localNamedTypes[typeID.TypeName] = t
			}
		}
	}
	// Find all locally defined enum values
	for _, _constDecl := range _decls.constDecls {
		for _, spec := range _constDecl.Specs {
			if spec.Type == nil {
				continue
			}
			if ident, ok := spec.Type.(*ast.Ident); ok && enums[ident.Name] != nil {
				enums[ident.Name].Values = append(enums[ident.Name].Values, enumVal{GenDecl: _constDecl.Decl, Decl: spec})
			}
		}

	}
	return ScanResult{
		Pkg:             pkg,
		Constants:       enums,
		MarkerCalls:     markerCalls,
		Interfaces:      interfaces,
		ConcreteTypes:   concreteTypes,
		SchemaMethods:   schemaMethods,
		SchemaFuncs:     schemaFuncs,
		LocalNamedTypes: localNamedTypes,
	}, nil
}

func loadPkgDecls(pkg *packages.Package) *decls {
	var (
		_decls decls
	)
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			switch _decl := decl.(type) {
			case *ast.FuncDecl:
				_decls.funcDecls = append(_decls.funcDecls, FuncDecl{
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
					_decls.typeDecls = append(_decls.typeDecls, TypeDecls{
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
					_decls.constDecls = append(_decls.constDecls, ConstDecls{
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
					_decls.varDecls = append(_decls.varDecls, VarDecls{
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
	return &_decls
}

func findNamedType(pkg *packages.Package, typeName string) (*types.Named, bool) {
	// Get the package's scope
	scope := pkg.Types.Scope()

	// Lookup the type by name
	obj := scope.Lookup(typeName)
	if obj == nil {
		return nil, false // Type not found
	}

	// Ensure the object is a TypeName
	typeNameObj, ok := obj.(*types.TypeName)
	if !ok {
		return nil, false // Not a named type
	}

	// Assert the type to *types.Named
	named, ok := typeNameObj.Type().(*types.Named)
	if !ok {
		return nil, false // Not a named type
	}

	return named, true
}
