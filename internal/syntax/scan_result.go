package syntax

import (
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"slices"
)

type PackageScanner interface {
	Scan(*packages.Package) (ScanResult, error)
}

// TypeDecl refers to the
type TypeDecl struct {
	Node  *dst.DeclStmt
	Pkg   *packages.Package
	File  *dst.File
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
	Identity *dst.Ident
	TypeInfo dst.Expr
	File     *dst.File
	Pkg      *packages.Package
	Decl     *TypeDecl
}

type TypeInfo interface {
	GetTypeName() string
	GetPkgPath() string
	GetTypeInfo() dst.Expr
	GetFile() *dst.File
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

func (t typeInfo) GetTypeInfo() dst.Expr {
	return t.TypeInfo
}

func (t typeInfo) GetFile() *dst.File {
	return t.File
}

func (t typeInfo) GetPkg() *packages.Package {
	return t.Pkg
}

type (
	TypeDecls struct {
		Pkg  *decorator.Package
		File *dst.File

		Decl  *dst.GenDecl
		Specs []*dst.TypeSpec
	}

	VarDecls struct {
		Pkg   *decorator.Package
		File  *dst.File
		Decl  *dst.GenDecl
		Specs []*dst.ValueSpec
	}

	SchemaMethod struct {
		Receiver   TypeID
		FuncName   string
		MarkerCall MarkerFunctionCall
	}
	SchemaFunction SchemaMethod
	// IfaceImplementations represents an interface type and its allowed types.
	// Error in the event that the loader encounters TypeName referenced on any
	// types declared outside of this package.
	// That is to say, in order for an interface implementations to work,
	// all supported references to it must be in the local package.
	IfaceImplementations struct {
		TypeSpec TypeSpec
		Impls    []TypeID
	}

	NamedTypeSpec struct {
		NamedType *types.Named
		TypeSpec  TypeSpec
		//Expr
	}

	EnumSet struct {
		TypeSpec TypeSpec
		Values   []ValueSpec
	}
)

func (s SchemaMethod) markerType() MarkerKind {
	return MarkerKindSchema
}

func (c EnumSet) markerType() MarkerKind {
	return MarkerKindEnum
}

func (i IfaceImplementations) markerType() MarkerKind {
	return MarkerKindInterface
}

// Marker is the common interface for all returned marker structs.
type Marker interface {
	markerType() MarkerKind
}

var (
	_ Marker = IfaceImplementations{}
	_ Marker = SchemaMethod{}
	_ Marker = EnumSet{}
)

// MarkerKind enumerates the four categories of markers we have.
type MarkerKind string

const (
	MarkerKindEnum      MarkerKind = "Enum"
	MarkerKindInterface MarkerKind = "InterfaceImpl"
	MarkerKindSchema    MarkerKind = "Schema"
)

//func (v VarDecls) importMap() syntax.ImportMap {
//	return v.File.Imports
//}

type VarDeclSet []VarConstDecl

func (vd VarDeclSet) MarkerFuncs() []MarkerFunctionCall {
	var result []MarkerFunctionCall
	for _, decl := range vd {
		for _, spec := range decl.Specs() {
			result = append(
				result,
				ParseValueExprForMarkerFunctionCall(spec)...,
			)
		}
	}
	return result
}

type decls struct {
	constDecls []VarConstDecl
	typeDecls  []TypeDecls
	varDecls   VarDeclSet
	funcDecls  []FuncDecl
}

type ScanResult struct {
	Pkg             *decorator.Package
	Constants       map[string]*EnumSet
	MarkerCalls     []MarkerFunctionCall
	Interfaces      map[string]IfaceImplementations
	ConcreteTypes   map[string]bool
	SchemaMethods   []SchemaMethod
	SchemaFuncs     []SchemaFunction
	LocalNamedTypes map[string]NamedTypeSpec
}

type seenPackages []string

func (s seenPackages) seen(pkg *decorator.Package) bool {
	return slices.Contains(s, pkg.ID)
}

func (s seenPackages) see(pkg *decorator.Package) seenPackages {
	return append(seenPackages{pkg.ID}, s...)
}

func (s seenPackages) add(pkg *decorator.Package) (seenPackages, bool) {
	if s.seen(pkg) {
		return s, false
	}
	return s.see(pkg), true
}

/**
 * Need a way to map out all the types in one go.  The structures here don't seem to do it.
 */
func LoadPackage(pkg *decorator.Package) (ScanResult, error) {
	return loadPackageInternal(pkg, seenPackages{})
}

func loadPackageInternal(pkg *decorator.Package, seen seenPackages) (ScanResult, error) {
	// Needs to discover:
	// 1. Enum (Const) Values
	// 2. Supported Interfaces
	// 3. Types to render
	var (
		_, ok           = seen.add(pkg)
		_decls          = loadPkgDecls(pkg)
		markerCalls     = _decls.varDecls.MarkerFuncs()
		enums           = map[string]*EnumSet{}
		interfaces      = map[string]IfaceImplementations{}
		concreteTypes   = map[string]bool{}
		schemaMethods   []SchemaMethod
		schemaFuncs     []SchemaFunction
		localNamedTypes = map[string]NamedTypeSpec{}
	)
	if !ok {
		return ScanResult{}, fmt.Errorf("circular package dependency detected. %v", seen)
	}
	for _, decl := range markerCalls {
		switch decl.CallExpr.MustIdentifyFunc().TypeName {
		case MarkerFuncNewEnumType:
			enums[decl.MustTypeArgument().TypeName] = &EnumSet{}
		case MarkerFuncNewInterfaceImpl:
			var (
				err   error
				iface = IfaceImplementations{}
			)
			if iface.Impls, err = decl.ParseTypesFromArgs(); err != nil {
				return ScanResult{}, err
			}
			interfaces[decl.MustTypeArgument().TypeName] = iface
			for _, impl := range iface.Impls {
				concreteTypes[impl.TypeName] = true
			}
		case MarkerFuncNewJSONSchemaMethod:
			method, err := decl.ParseSchemaMethod()
			concreteTypes[method.Receiver.TypeName] = true
			if err != nil {
				return ScanResult{}, err
			}
			schemaMethods = append(schemaMethods, method)
		case MarkerFuncNewJSONSchemaBuilder:
			method, err := decl.ParseSchemaFunc()
			concreteTypes[method.Receiver.TypeName] = true
			if err != nil {
				return ScanResult{}, err
			}
			schemaFuncs = append(schemaFuncs, method)

		default:
			return ScanResult{}, fmt.Errorf("unsupported marker function: %s", decl.CallExpr.MustIdentifyFunc())
		}
	}

	for _, _typeDecl := range _decls.typeDecls {
		for _, spec := range _typeDecl.Specs {
			var (
				typeID = TypeID{PkgPath: pkg.PkgPath, TypeName: spec.Name.Name}
			)
			if iface, ok := interfaces[typeID.TypeName]; ok {
				iface.TypeSpec = NewTypeSpec(_typeDecl.Decl, spec, _typeDecl.Pkg, _typeDecl.File)
				interfaces[typeID.TypeName] = iface
			} else if enum, ok := enums[typeID.TypeName]; ok {
				enum.TypeSpec = NewTypeSpec(_typeDecl.Decl, spec, _typeDecl.Pkg, _typeDecl.File)
			} else {
				var t = NamedTypeSpec{
					TypeSpec: NewTypeSpec(_typeDecl.Decl, spec, _typeDecl.Pkg, _typeDecl.File),
				}
				if t.NamedType, ok = findNamedType(_typeDecl.Pkg, spec.Name.Name); !ok {
					return ScanResult{}, fmt.Errorf("unable to load named type for %s declared at %s", spec.Name.Name, t.TypeSpec.Position())
				}
				localNamedTypes[typeID.TypeName] = t
			}
		}
	}
	// Find all locally defined enum values
	for _, _constDecl := range _decls.constDecls {
		for _, spec := range _constDecl.Specs() {
			if !spec.HasType() {
				continue
			}
			if ident, ok := spec.Type().(*dst.Ident); ok {
				typeID := TypeID{TypeName: ident.Name}
				if ident.Path == "" {
					typeID.PkgPath = pkg.PkgPath
				} else {
					typeID.PkgPath = ident.Path
				}
				enums[typeID.TypeName].Values = append(enums[typeID.TypeName].Values, spec)
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

func loadPkgDecls(pkg *decorator.Package) *decls {
	var (
		_decls decls
	)
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			switch _decl := decl.(type) {
			case *dst.FuncDecl:
				_decls.funcDecls = append(_decls.funcDecls, NewFuncDecl(_decl, pkg, file))
			case *dst.GenDecl:
				switch _decl.Tok {
				case token.TYPE:
					var specs []*dst.TypeSpec
					for _, spec := range _decl.Specs {
						specs = append(specs, spec.(*dst.TypeSpec))
					}
					_decls.typeDecls = append(_decls.typeDecls, TypeDecls{
						Pkg:   pkg,
						File:  file,
						Decl:  _decl,
						Specs: specs,
					})
				case token.CONST:
					_decls.constDecls = append(_decls.constDecls, NewVarConstDecl(_decl, pkg, file))
				case token.VAR:
					_decls.varDecls = append(_decls.varDecls, NewVarConstDecl(_decl, pkg, file))
				default:
				}
			}
		}
	}
	return &_decls
}

func (n NamedTypeSpec) Position() token.Position {
	return n.TypeSpec.Position()
}

func findNamedType(pkg *decorator.Package, typeName string) (*types.Named, bool) {
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
