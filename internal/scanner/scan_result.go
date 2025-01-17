package scanner

import (
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/tylergannon/go-gen-jsonschema/internal/common"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
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
	Implementations []common.TypeID
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

	FuncDecl struct {
		Pkg  *decorator.Package
		File *dst.File
		Decl *dst.FuncDecl
	}

	ConstDecls struct {
		Pkg   *decorator.Package
		File  *dst.File
		Decl  *dst.GenDecl
		Specs []*dst.ValueSpec
	}

	SchemaMethod struct {
		Receiver   common.TypeID
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
		Pkg      *decorator.Package
		File     *dst.File
		TypeID   common.TypeID
		Impls    []common.TypeID
		Position token.Position
	}

	enumVal struct {
		GenDecl *dst.GenDecl
		Decl    *dst.ValueSpec
	}

	AnyTypeSpec struct {
		Spec dst.Expr
		File *dst.File
		Pkg  *decorator.Package
	}

	NamedTypeSpec struct {
		NamedType *types.Named
		GenDecl   *dst.GenDecl
		TypeSpec  *dst.TypeSpec
		AnyTypeSpec
	}

	EnumSet struct {
		GenDecl  *dst.GenDecl
		TypeSpec *dst.TypeSpec
		Pkg      *decorator.Package
		File     *dst.File
		TypeID   common.TypeID
		Values   []enumVal
	}
)

func (n NamedTypeSpec) GetDescription() string {
	if len(n.TypeSpec.Decorations().Start.All()) > 0 {
		return BuildComments(n.TypeSpec.Decorations())
	} else if len(n.GenDecl.Specs) == 1 && len(n.GenDecl.Decs.Start.All()) > 0 {
		return BuildComments(n.GenDecl.Decorations())
	}
	return ""
}

func (a AnyTypeSpec) Derive(spec dst.Expr) AnyTypeSpec {
	return AnyTypeSpec{
		Spec: spec,
		File: a.File,
		Pkg:  a.Pkg,
	}
}

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
	Pkg             *decorator.Package
	Constants       map[common.TypeID]*EnumSet
	MarkerCalls     []MarkerFunctionCall
	Interfaces      map[string]IfaceImplementations
	ConcreteTypes   map[common.TypeID]bool
	SchemaMethods   []SchemaMethod
	SchemaFuncs     []SchemaFunction
	LocalNamedTypes map[string]NamedTypeSpec
}

func LoadPackage(pkg *decorator.Package) (ScanResult, error) {
	// Needs to discover:
	// 1. Enum (Const) Values
	// 2. Supported Interfaces
	// 3. Types to render
	var (
		_decls          = loadPkgDecls(pkg)
		markerCalls     = _decls.varDecls.MarkerFuncs()
		enums           = map[common.TypeID]*EnumSet{}
		interfaces      = map[string]IfaceImplementations{}
		concreteTypes   = map[common.TypeID]bool{}
		schemaMethods   []SchemaMethod
		schemaFuncs     []SchemaFunction
		localNamedTypes = map[string]NamedTypeSpec{}
	)
	for _, decl := range markerCalls {
		switch decl.Function {
		case MarkerFuncNewEnumType:
			enums[decl.TypeArgument.Localize(pkg.PkgPath)] = &EnumSet{
				Pkg:    decl.Pkg,
				TypeID: decl.TypeArgument.Localize(pkg.PkgPath),
			}
		case MarkerFuncNewInterfaceImpl:
			var (
				err   error
				iface = IfaceImplementations{
					TypeID:   *decl.TypeArgument,
					Position: decl.Position,
				}
			)
			if iface.Impls, err = decl.ParseTypesFromArgs(); err != nil {
				return ScanResult{}, err
			}
			interfaces[iface.TypeID.TypeName] = iface
			for _, impl := range iface.Impls {
				concreteTypes[impl.Concrete()] = true
			}
		case MarkerFuncNewJSONSchemaMethod:
			method, err := decl.ParseSchemaMethod()
			concreteTypes[method.Receiver.Concrete()] = true
			if err != nil {
				return ScanResult{}, err
			}
			schemaMethods = append(schemaMethods, method)
		case MarkerFuncNewJSONSchemaBuilder:
			method, err := decl.ParseSchemaFunc()
			concreteTypes[method.Receiver.Concrete()] = true
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
				typeID = common.TypeID{DeclaredLocally: true, TypeName: spec.Name.Name}
			)
			if iface, ok := interfaces[typeID.TypeName]; ok {
				iface.Pkg = pkg
				iface.File = _typeDecl.File
				interfaces[typeID.TypeName] = iface
			} else if enum, ok := enums[typeID.Concrete().Localize(pkg.PkgPath)]; ok {
				enum.GenDecl = _typeDecl.Decl
				enum.TypeSpec = spec
				enum.File = _typeDecl.File
			} else {
				var t = NamedTypeSpec{
					GenDecl:  _typeDecl.Decl,
					TypeSpec: spec,
					AnyTypeSpec: AnyTypeSpec{
						File: _typeDecl.File,
						Pkg:  _typeDecl.Pkg,
						Spec: spec.Type,
					},
				}
				if t.NamedType, ok = findNamedType(_typeDecl.Pkg, spec.Name.Name); !ok {
					pos := NodePosition(_typeDecl.Pkg, spec)
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
			if ident, ok := spec.Type.(*dst.Ident); ok {
				typeID := common.TypeID{TypeName: ident.Name}
				if ident.Path == "" {
					typeID.PkgPath = pkg.PkgPath
				} else {
					typeID.PkgPath = ident.Path
				}
				enums[typeID].Values = append(enums[typeID].Values, enumVal{GenDecl: _constDecl.Decl, Decl: spec})
			}
			//if ident, ok := spec.Type.(*dst.Ident); ok && enums[ident.Name] != nil {
			//	enums[ident.Name].Values = append(enums[ident.Name].Values, enumVal{GenDecl: _constDecl.Decl, Decl: spec})
			//}
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
				_decls.funcDecls = append(_decls.funcDecls, FuncDecl{
					Pkg:  pkg,
					File: file,
					Decl: _decl,
				})
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
					var values []*dst.ValueSpec
					for _, spec := range _decl.Specs {
						values = append(values, spec.(*dst.ValueSpec))
					}
					_decls.constDecls = append(_decls.constDecls, ConstDecls{
						Pkg:   pkg,
						File:  file,
						Decl:  _decl,
						Specs: values,
					})
				case token.VAR:
					var specs []*dst.ValueSpec
					for _, spec := range _decl.Specs {
						specs = append(specs, spec.(*dst.ValueSpec))
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

func (n NamedTypeSpec) Position() token.Position {
	return NodePosition(n.Pkg, n.TypeSpec)
}

func (a AnyTypeSpec) Position() token.Position {
	return NodePosition(a.Pkg, a.Spec)
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
