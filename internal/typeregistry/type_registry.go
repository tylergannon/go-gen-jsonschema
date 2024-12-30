package typeregistry

import (
	"errors"
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"go/token"
	"go/types"
	"log"
	"path/filepath"
	"strings"
)

var (
	ErrUnsupportedType = errors.New("unsupported type")
)

type EnumEntry struct {
	Name        string
	Value       string
	Decorations *dst.NodeDecs
}

type Registry struct {
	packages       map[string]*decorator.Package
	typeMap        map[TypeID]*typeSpec
	unionTypes     map[TypeID]*UnionTypeDecl
	interfaceTypes map[TypeID]*InterfaceTypeDecl
	imports        map[string]*decorator.Package
	constants      map[TypeID][]EnumEntry
	funcs          map[TypeID]*FuncEntry
	unmarshalers   map[TypeID]*FuncEntry
}

type TypeID string

func (id TypeID) shorten(pkg *decorator.Package) TypeID {
	if len(id) < 30 || strings.HasPrefix(string(id), "struct") {
		return id
	}
	//s := string(ID)
	rel, err := filepath.Rel(pkg.PkgPath, string(id))
	if err != nil {
		return id
	}
	//s = strings.ReplaceAll(s, Pkg.PkgPath, "")
	return TypeID(rel)
}

type IdentifiableByType interface {
	// ID returns a string name composed of package and type name.
	ID() TypeID
}

func NewTypeID(pkgPath, typeName string) TypeID {
	if typeName == "" {
		panic("typeName is empty")
	}
	if pkgPath == "" {
		panic("pkgPath is empty")
	}
	return TypeID(pkgPath + "." + typeName)
}

func (r *Registry) GetTypeByName(name, pkgPath string) (TypeSpec, bool) {
	t, ok := r.getType(name, pkgPath)
	if !ok {
		fmt.Println("Not found: ", name, pkgPath)
		for k := range r.typeMap {
			fmt.Println(k)
		}
	}
	return t, ok
}

func (r *Registry) loadUnionTypeAlternates(alt *UnionTypeDecl) (specs []*AlternativeTypeSpec) {
	for _, typeAlt := range alt.Alternatives {
		var altPkgPath, altTypeName string
		if err := r.LoadAndScan(typeAlt.ImportMap[""]); err != nil {
			panic(err)
		}
		conversionFuncTypeID := NewTypeID(typeAlt.ImportMap[""], typeAlt.ConversionFunc)
		if funcEntry := r.funcs[conversionFuncTypeID]; funcEntry != nil {
			altPkgPath, altTypeName = funcEntry.ReceiverOrArgType()
		} else {
			panic("couldn't find func " + string(conversionFuncTypeID))
		}

		if err := r.LoadAndScan(altPkgPath); err != nil {
			panic(err)
		}
		sourceTypeID := NewTypeID(altPkgPath, altTypeName)

		if altTS, ok := r.typeMap[sourceTypeID]; !ok {
			panic("couldn't find type alt " + string(sourceTypeID) + " for func " + typeAlt.ConversionFunc)
		} else {
			specs = append(specs, &AlternativeTypeSpec{
				TypeSpec:       altTS,
				ConversionFunc: typeAlt.ConversionFunc,
			})
		}
	}
	return specs
}

func (r *Registry) getType(name string, pkgPath string) (*typeSpec, bool) {
	if r.packages[pkgPath] == nil {
		if err := r.LoadAndScan(pkgPath); err != nil {
			fmt.Println(err)
			return nil, false
		}
	}
	typeID := NewTypeID(pkgPath, name)

	if ts, ok := r.typeMap[typeID]; ok {
		// If the requested type is known as a union type, load the type alternatives
		// on the typeSpec.
		if alt, ok := r.unionTypes[typeID]; ok && len(ts.alts) == 0 {
			ts.alts = r.loadUnionTypeAlternates(alt)
		} else if iface, ok := r.interfaceTypes[typeID]; ok {
			ts.alts = r.loadInterfaceImplementations(iface)
		}
		return ts, true
	}
	return nil, false
}

func (r *Registry) registerConstDecl(file *dst.File, pkg *decorator.Package, decl *dst.GenDecl) error {
	for _, spec := range decl.Specs {
		v := spec.(*dst.ValueSpec)

		typeIdent, ok := v.Type.(*dst.Ident)
		if !ok {
			continue
		}
		typeID := NewTypeID(pkg.PkgPath, typeIdent.Name)
		name := v.Names[0].Name
		value := v.Values[0]
		lit, ok := value.(*dst.BasicLit)

		if !ok || lit.Kind != token.STRING {
			continue
		}

		r.constants[typeID] = append(r.constants[typeID], EnumEntry{
			Name:        name,
			Value:       strings.TrimSuffix(strings.TrimPrefix(lit.Value, "\""), "\""),
			Decorations: v.Decorations(),
		})
	}
	return nil
}

func (r *Registry) loadInterfaceImplementations(iface *InterfaceTypeDecl) (specs []*AlternativeTypeSpec) {
	for _, typeAlt := range iface.Implementations {
		if err := r.LoadAndScan(typeAlt.PkgPath); err != nil {
			panic(err)
		}
		sourceTypeID := NewTypeID(typeAlt.PkgPath, typeAlt.TypeName)

		if altTS, ok := r.typeMap[sourceTypeID]; !ok {
			panic("couldn't find interface implementation " + string(sourceTypeID) + " for interface " + iface.InterfaceTypeName)
		} else {
			specs = append(specs, &AlternativeTypeSpec{
				TypeSpec:    altTS,
				PointerType: typeAlt.IsPointer,
			})
		}
	}
	return specs
}

type (
	BasicType   string
	InvalidType string
)

func (BasicType) IsInterface() bool {
	return false
}

func (b BasicType) GetType() types.Type {
	//TODO implement me
	panic("implement me")
}

func (b BasicType) GetTypeSpec() *dst.TypeSpec {
	//TODO implement me
	panic("implement me")
}

func (b BasicType) Pkg() *decorator.Package {
	//TODO implement me
	panic("implement me")
}

func (b BasicType) GenDecl() *dst.GenDecl {
	//TODO implement me
	panic("implement me")
}

func (b BasicType) File() *dst.File {
	//TODO implement me
	panic("implement me")
}

func (b BasicType) Decorations() *dst.NodeDecs {
	//TODO implement me
	panic("implement me")
}

func (b BasicType) ID() TypeID {
	return TypeID(string(b))
}

func (b BasicType) Alternatives() []*AlternativeTypeSpec {
	return nil
}

const (
	BasicTypeString BasicType = "string"
	BasicTypeInt    BasicType = "int"
	BasicTypeBool   BasicType = "bool"
	BasicTypeFloat  BasicType = "float"
)

var _ TypeSpec = BasicTypeString

const (
	InvalidTypeChannel   InvalidType = "channel"
	InvalidTypeFunc      InvalidType = "func"
	InvalidTypeInterface InvalidType = "interface"
	InvalidTypeUnsafe    InvalidType = "unsafe"
)

func (i InvalidType) ID() TypeID {
	return TypeID(string(i))
}

type TypeSpec interface {
	IdentifiableByType
	GetTypeSpec() *dst.TypeSpec
	Pkg() *decorator.Package
	GenDecl() *dst.GenDecl
	File() *dst.File
	Decorations() *dst.NodeDecs
	GetType() types.Type
	Alternatives() []*AlternativeTypeSpec
	IsInterface() bool
}

type AlternativeTypeSpec struct {
	TypeSpec       TypeSpec
	ConversionFunc string
	PointerType    bool
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

type InterfaceTypeDecl struct {
	importMap            ImportMap
	InterfacePackagePath string
	InterfaceTypeName    string
	Implementations      []InterfaceImpl
	File                 *dst.File          `json:"-"`
	Pkg                  *decorator.Package `json:"-"`
}

func (d *InterfaceTypeDecl) ID() TypeID {
	return NewTypeID(d.InterfacePackagePath, d.InterfaceTypeName)
}

// SetTypeAlternativeDecl interprets the index expression for the identity of the
// type alt.
// TODO: I don't think we should be returning anything in the default case below.
// anybody know why I did that?
func SetTypeAlternativeDecl(importMap ImportMap, expr *dst.IndexExpr) *UnionTypeDecl {
	switch index := expr.Index.(type) {
	case *dst.Ident:
		//log.Printf("Name: %s, Path: %s, Obj: %v, %T", nodeImpl.Name, nodeImpl.Path, nodeImpl.Obj, nodeImpl.Obj)
		return &UnionTypeDecl{
			importMap:           importMap,
			DestTypePackagePath: importMap[""],
			DestTypeName:        index.Name,
		}
	default:
		log.Printf("Expr: %T, %v", index, index)
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
	alts    []*AlternativeTypeSpec
}

func (ts *typeSpec) Alternatives() []*AlternativeTypeSpec {
	return ts.alts
}

func (ts *typeSpec) GetType() types.Type {
	return ts.pkg.Types.Scope().Lookup(ts.typeSpec.Name.Name).(*types.TypeName).Type()
}

func (ts *typeSpec) IsInterface() bool {
	_, ok := ts.GetType().Underlying().(*types.Interface)
	return ok
}

func (ts *typeSpec) Decorations() *dst.NodeDecs {
	var nodeDecs dst.NodeDecs
	if start := ts.typeSpec.Decorations().Start; start != nil {
		nodeDecs.Start.Replace(start...)
	} else if start := ts.genDecl.Decorations().Start; start != nil {
		nodeDecs.Start.Replace(start...)
	}
	if end := ts.typeSpec.Decorations().End; end != nil {
		nodeDecs.End.Replace(end...)
	} else if end := ts.genDecl.Decorations().End; end != nil {
		nodeDecs.End.Replace(end...)
	}
	return &nodeDecs
}

func (ts *typeSpec) ID() TypeID {
	return NewTypeID(ts.pkg.PkgPath, ts.typeSpec.Name.Name)
}

var _ TypeSpec = (*typeSpec)(nil)

func (ts *typeSpec) GetTypeSpec() *dst.TypeSpec {
	return ts.typeSpec
}

func (ts *typeSpec) Pkg() *decorator.Package {
	if ts == nil {
		fmt.Println("ts is frikkin nil!")
	} else if ts.pkg == nil {
		inspect("ts", ts.typeSpec)
	}
	return ts.pkg
}

func (ts *typeSpec) GenDecl() *dst.GenDecl {
	return ts.genDecl
}

func (ts *typeSpec) File() *dst.File {
	return ts.file
}
