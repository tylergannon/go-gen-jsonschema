package syntax

import (
	"fmt"
	"go/token"
	"slices"
	"strings"
	"unicode"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/tylergannon/structtag"
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
		Receiver         TypeID
		SchemaMethodName string
		MarkerCall       MarkerFunctionCall
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

	EnumSet struct {
		TypeSpec TypeSpec
		Values   []ValueSpec
	}
)

func (s SchemaMethod) IsPointer() bool {
	return s.Receiver.Indirection == Pointer
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
	localTypeNames  map[string]bool
	SchemaMethods   []SchemaMethod
	SchemaFuncs     []SchemaFunction
	LocalNamedTypes map[string]TypeSpec
	remoteTypes     typesMap
	deps            map[string]ScanResult
	// temp variable used during resolution only.
	resolveQueue            []TypeSpec
	alreadyTraversedLocally map[string]bool
}

func (s ScanResult) GetPackage(pkgPath string) (ScanResult, bool) {
	if pkgPath == s.Pkg.PkgPath {
		return s, true
	}
	res, ok := s.deps[pkgPath]
	return res, ok
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

// LoadPackage is the main entry point that creates a ScanResult for the given package.
// Note: we pass a non-nil map to loadPackageInternal(...) so we can safely store references
// to local types without panicking.
func LoadPackage(pkg *decorator.Package) (res ScanResult, err error) {
	res = newScanResult(pkg, map[string]ScanResult{})
	// Pass an empty map so we never do `typesToMap[foo] = true` on a nil map.
	err = res.loadPackageInternal(seenPackages{}, make(map[string]bool))
	return
}

func loadPackageForTest(pkg *decorator.Package, typesToInclude ...string) (ScanResult, error) {
	var types = make(map[string]bool)
	for _, typeName := range typesToInclude {
		types[typeName] = true
	}
	scanResult := newScanResult(pkg, map[string]ScanResult{})
	err := scanResult.loadPackageInternal(seenPackages{}, types)
	return scanResult, err
}

func newScanResult(pkg *decorator.Package, deps map[string]ScanResult) ScanResult {
	return ScanResult{
		Pkg:                     pkg,
		Constants:               make(map[string]*EnumSet),
		MarkerCalls:             make([]MarkerFunctionCall, 0),
		Interfaces:              make(map[string]IfaceImplementations),
		SchemaMethods:           make([]SchemaMethod, 0),
		SchemaFuncs:             make([]SchemaFunction, 0),
		LocalNamedTypes:         make(map[string]TypeSpec),
		remoteTypes:             typesMap{},
		localTypeNames:          make(map[string]bool),
		deps:                    deps,
		alreadyTraversedLocally: make(map[string]bool),
	}
}

type typesMap map[string]map[string]bool

func (t typesMap) addTypeByID(typ TypeID) {
	t.addType(typ.PkgPath, typ.TypeName)
}

func (t typesMap) addType(pkgPath, typeName string) {
	if t[pkgPath] == nil {
		t[pkgPath] = map[string]bool{typeName: true}
	} else {
		t[pkgPath][typeName] = true
	}
}

func (r *ScanResult) resolveTypeLocal(name string) (Expr, error) {
	fmt.Printf("resolveTypeLocal(%s)\n", name)
	var namedTypeNames []string
	for typeName := range r.LocalNamedTypes {
		namedTypeNames = append(namedTypeNames, typeName)
	}
	fmt.Printf("Named types in %s: %v\n", r.Pkg.PkgPath, namedTypeNames)
	t, ok := r.LocalNamedTypes[name]
	if !ok {
		return nil, fmt.Errorf("type %s not found", name)
	}
	if ident, ok := t.Type().Expr().(*dst.Ident); ok {
		return r.resolveType(r.Pkg.PkgPath, IdentExpr{STExpr: NewExpr(ident, t.Pkg(), t.File())})
	}
	return t.Type(), nil
}

func (r *ScanResult) resolveTypeRemote(path, name string) (Expr, error) {
	fmt.Printf("Resolving remote type %s from %s\n", name, path)
	scanResult, ok := r.GetPackage(path)
	if !ok {
		return nil, fmt.Errorf("package %s not found", path)
	}
	var namedTypeNames []string
	for typeName := range scanResult.LocalNamedTypes {
		namedTypeNames = append(namedTypeNames, typeName)
	}
	fmt.Printf("resolveTypeRemote: Named types in %s: %v\n", path, namedTypeNames)
	return scanResult.resolveTypeLocal(name)
}

func (r *ScanResult) resolveType(pkgPath string, ident IdentExpr) (Expr, error) {
	fmt.Printf("resolveType %s: Resolving type %v\n", r.Pkg.PkgPath, ident)
	e := ident.Concrete
	if e.Path == "" {
		s, ok := r.GetPackage(pkgPath)
		if !ok {
			return nil, fmt.Errorf("package %s not found", pkgPath)
		}
		fmt.Printf("resolveType(pkgPath=%s, %s.%s)\n", pkgPath, e.Path, e.Name)
		return s.resolveTypeLocal(e.Name)
	} else {
		fmt.Printf("Resolving remote type %s from %s\n", e.Name, e.Path)
		return r.resolveTypeRemote(e.Path, e.Name)
	}
}

func (r *ScanResult) loadPackageInternal(seen seenPackages, typesToMap map[string]bool) error {
	fmt.Printf("Loading package %s with typesToMap %v\n", r.Pkg.PkgPath, typesToMap)
	if typesToMap == nil {
		// Safety check in case it's ever passed nil from some other call site
		typesToMap = make(map[string]bool)
	}

	var (
		_, ok  = seen.add(r.Pkg)
		_decls = loadPkgDecls(r.Pkg)
	)
	if !ok {
		return fmt.Errorf("circular package dependency detected. %v", seen)
	}

	r.MarkerCalls = _decls.varDecls.MarkerFuncs()
	for _, decl := range r.MarkerCalls {
		switch decl.CallExpr.MustIdentifyFunc().TypeName {
		case MarkerFuncNewEnumType:
			r.Constants[decl.MustTypeArgument().TypeName] = &EnumSet{}
		case MarkerFuncNewInterfaceImpl:
			var (
				err   error
				iface = IfaceImplementations{}
			)
			if iface.Impls, err = decl.ParseTypesFromArgs(); err != nil {
				return err
			}
			r.Interfaces[decl.MustTypeArgument().TypeName] = iface
			for _, impl := range iface.Impls {
				if impl.PkgPath == r.Pkg.PkgPath {
					r.localTypeNames[impl.TypeName] = true
					typesToMap[impl.TypeName] = true
				} else {
					r.remoteTypes.addTypeByID(impl)
				}
			}
		case MarkerFuncNewJSONSchemaMethod:
			method, err := decl.ParseSchemaMethod()
			if err != nil {
				return err
			}
			r.localTypeNames[method.Receiver.TypeName] = true
			typesToMap[method.Receiver.TypeName] = true
			r.SchemaMethods = append(r.SchemaMethods, method)
		case MarkerFuncNewJSONSchemaBuilder:
			fn, err := decl.ParseSchemaFunc()
			if err != nil {
				return err
			}
			r.localTypeNames[fn.Receiver.TypeName] = true
			typesToMap[fn.Receiver.TypeName] = true
			r.SchemaFuncs = append(r.SchemaFuncs, fn)

		default:
			return fmt.Errorf("unsupported marker function: %s", decl.CallExpr.MustIdentifyFunc())
		}
	}

	for _, _typeDecl := range _decls.typeDecls {
		for _, spec := range _typeDecl.Specs {
			var (
				typeID = TypeID{PkgPath: r.Pkg.PkgPath, TypeName: spec.Name.Name}
			)
			if iface, ok := r.Interfaces[typeID.TypeName]; ok {
				iface.TypeSpec = NewTypeSpec(_typeDecl.Decl, spec, _typeDecl.Pkg, _typeDecl.File)
				r.Interfaces[typeID.TypeName] = iface
			} else if enum, ok := r.Constants[typeID.TypeName]; ok {
				enum.TypeSpec = NewTypeSpec(_typeDecl.Decl, spec, _typeDecl.Pkg, _typeDecl.File)
			} else {
				r.LocalNamedTypes[typeID.TypeName] = NewTypeSpec(_typeDecl.Decl, spec, _typeDecl.Pkg, _typeDecl.File)
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
					typeID.PkgPath = r.Pkg.PkgPath
				} else {
					typeID.PkgPath = ident.Path
				}
				// Only append to the enum set if r.Constants[typeID.TypeName] is non-nil:
				if e, exists := r.Constants[typeID.TypeName]; exists && e != nil {
					e.Values = append(e.Values, spec)
				}
			}
		}
	}

	for typeName := range typesToMap {
		if _, ok := r.LocalNamedTypes[typeName]; !ok {
			if r.Constants[typeName] != nil {
				continue
			}
			if _, ok = r.Interfaces[typeName]; ok {
				continue
			}
			return fmt.Errorf("undeclared local type found: %s", typeName)
		} else {
			// We'll queue it for resolution
			r.resolveQueue = append(r.resolveQueue, r.LocalNamedTypes[typeName])
		}
	}

	if err := r.resolveTypes(); err != nil {
		return err
	}
	return nil
}

func (r *ScanResult) resolveTypeExpr(_expr Expr, seen SeenTypes) error {
	switch expr := _expr.Expr().(type) {
	case *dst.ParenExpr:
		return r.resolveTypeExpr(_expr.NewExpr(expr.X), seen)
	case *dst.StarExpr:
		return r.resolveTypeExpr(_expr.NewExpr(expr.X), seen)
	case *dst.SliceExpr:
		return r.resolveTypeExpr(_expr.NewExpr(expr.X), seen)
	case *dst.ArrayType:
		return r.resolveTypeExpr(_expr.NewExpr(expr.Elt), seen)
	case *dst.StructType:
		for _, field := range expr.Fields.List {
			if skipField(field) {
				continue
			}
			if err := r.resolveTypeExpr(_expr.NewExpr(field.Type), seen); err != nil {
				return fmt.Errorf("struct field at %s: %w", _expr.NewExpr(field.Type).Position(), err)
			}
		}
	case *dst.Ident:
		if expr.Path == "" || expr.Path == r.Pkg.PkgPath {
			// It's either a basic type or a locally-defined named type
			if BasicTypes[expr.Name] {
				return nil // basic type
			}
			if named, ok := r.LocalNamedTypes[expr.Name]; !ok {
				if r.Constants[expr.Name] != nil {
					return nil
				} else if _, ok := r.Interfaces[expr.Name]; !ok {
					return nil
				}
				return fmt.Errorf("undeclared local %s type found: %s at %s", expr.Name, _expr.Details(), _expr.Position())
			} else {
				var added bool
				seen, added = seen.Add(named.ID())
				if !added {
					return fmt.Errorf("cyclic dependency found at %s", named.Position())
				}
				if r.alreadyTraversedLocally[expr.Name] {
					return nil
				}
				if err := r.resolveTypeExpr(named.Type(), seen); err != nil {
					return err
				}
				r.alreadyTraversedLocally[expr.Name] = true
			}
		} else {
			r.remoteTypes.addType(expr.Path, expr.Name)
		}
	case *dst.SelectorExpr:
		// Instead of panicking, at least store the remote reference
		if xIdent, ok := expr.X.(*dst.Ident); ok {
			if xIdent.Path != "" {
				r.remoteTypes.addType(xIdent.Path, expr.Sel.Name)
			} else {
				// Fallback if there's no path, treat the 'X' as the package name
				r.remoteTypes.addType(xIdent.Name, expr.Sel.Name)
			}
		} else {
			return fmt.Errorf("unhandled selector expression: %s", _expr.Details())
		}
	case *dst.BasicLit:
		return nil
	default:
		return fmt.Errorf("unhandled expression %s", _expr.Details())
	}
	return nil
}

func skipField(field *dst.Field) bool {
	if len(field.Names) == 0 { // don't skip embedded
		return false
	}
	if idx := slices.IndexFunc(field.Names, func(ident *dst.Ident) bool {
		return unicode.IsLower(rune(ident.Name[0]))
	}); idx == -1 {
		return true // skip if all names are lowercased (non-exported)
	}
	if field.Tag == nil {
		return false
	}
	tags, _ := structtag.Parse(strings.Trim(field.Tag.Value, "`"))
	jsonTag, _ := tags.Get("json")
	return len(jsonTag.Options) == 0 || jsonTag.Options[0] == "-"
}

func (r *ScanResult) resolveTypes() error {
	var (
		ts  TypeSpec
		err error
	)
	for len(r.resolveQueue) > 0 {
		ts = r.resolveQueue[0]
		r.resolveQueue = r.resolveQueue[1:]
		if r.alreadyTraversedLocally[ts.Concrete.Name.Name] {
			continue
		}
		// Pass a non-nil "seen" so we can detect cycles properly.
		if err = r.resolveTypeExpr(NewExpr(ts.Concrete.Type, ts.pkg, ts.file), nil); err != nil {
			return fmt.Errorf("resolving Concrete at %s: %w", ts.Position(), err)
		}
	}
	for pkgPath, typeNames := range r.remoteTypes {
		fmt.Printf("Resolving remote types %v for %s\n", typeNames, pkgPath)
		if remote, ok := r.deps[pkgPath]; ok {
			for typeName := range typeNames {
				if !remote.alreadyTraversedLocally[typeName] {
					remote.resolveQueue = append(remote.resolveQueue, remote.LocalNamedTypes[typeName])
				}
			}
			if err = remote.resolveTypes(); err != nil {
				return fmt.Errorf("resolving type at %s: %w", pkgPath, err)
			}
		} else if pkgs, err := Load(pkgPath); err != nil {
			return err
		} else {
			remote = newScanResult(pkgs[0], r.deps)
			if err = remote.loadPackageInternal(seenPackages{}, typeNames); err != nil {
				return fmt.Errorf("resolving type at %s: %w", pkgPath, err)
			}
			r.deps[pkgPath] = remote
		}
	}
	return nil
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

var BasicTypes = map[string]bool{
	"int":        true,
	"int8":       true,
	"int16":      true,
	"int32":      true,
	"int64":      true,
	"uint":       true,
	"uint8":      true,
	"uint16":     true,
	"uint32":     true,
	"uint64":     true,
	"uintptr":    true,
	"string":     true,
	"bool":       true,
	"float32":    true,
	"float64":    true,
	"complex64":  true,
	"complex128": true,
	"byte":       true, // alias for uint8
	"rune":       true, // alias for int32
	"error":      true,
}
