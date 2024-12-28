package typeregistry

import (
	"errors"
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/tylergannon/go-gen-jsonschema/internal/loader"
	"go/token"
	"go/types"
)

const (
	JSONSchemaPackage = "github.com/tylergannon/go-gen-jsonschema"
	// UnionTypeFunc Used for denoting basic alts for a struct type.  These are special because each "type alt"
	// needs to have a converter func that renders it into the destination type.
	UnionTypeFunc = "SetTypeAlternative"
	// ImplementationFunc denotes the list of acceptable implementations for an interface type.
	// These are different from the "type alt" in that there is no conversion function.  By being
	// an implementation of the denoted interface, they are directly assignable to that type.
	ImplementationFunc = "SetImplementations"
	TypeAltFunc        = "Alt"
)

func NewRegistry(pkgs []*decorator.Package) (*Registry, error) {
	registry := &Registry{
		typeMap:    map[TypeID]*typeSpec{},
		packages:   map[string]*decorator.Package{},
		unionTypes: map[TypeID]*UnionTypeDecl{},
		imports:    map[string]*decorator.Package{},
		constants:  map[TypeID][]EnumEntry{},
		funcs:      map[TypeID]*FuncEntry{},
	}
	for _, pkg := range pkgs {
		if err := registry.scan(pkg); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

// LoadAndScan finds all type declarations and type alternate declarations.
// Does *NOT* differentiate type declarations, filter them, validate them, etc.
func (r *Registry) LoadAndScan(pkgPath string) error {
	if r.packages[pkgPath] != nil {
		return nil
	}
	var (
		pkgs []*decorator.Package
		err  error
	)
	if pkg, ok := r.imports[pkgPath]; ok {
		pkgs = append(pkgs, pkg)
	} else if pkgs, err = loader.Load(pkgPath); err != nil {
		return err
	}
	for _, pkg := range pkgs {
		if err = r.scan(pkg); err != nil {
			return err
		}
	}
	return nil
}

func (r *Registry) scan(pkg *decorator.Package) error {
	if r.packages[pkg.PkgPath] != nil {
		return nil
	}
	funcs := map[TypeID]*FuncEntry{}

	for _, file := range pkg.Syntax {
		if funcsTemp, err := r.scanFile(file, pkg); err != nil {
			return err
		} else {
			for id, funcDecl := range funcsTemp {
				funcs[id] = funcDecl
			}
		}
	}

	r.packages[pkg.PkgPath] = pkg

	for _, obj := range pkg.TypesInfo.Defs {
		if obj == nil {
			continue
		}

		funcObj, ok := obj.(*types.Func)
		if !ok {
			continue
		}
		typeID := NewTypeID(pkg.PkgPath, funcNameFromTypes(funcObj))
		if f, ok := funcs[typeID]; ok {
			f.Func = funcObj
			f.typeID = typeID
			r.funcs[typeID] = f
		}
	}
	return nil
}

// funcNameFromDst returns the name of the function if the function takes no
// receiver.  If it takes a receiver, the function will be namespaced with the
// type name of the receiver.
func funcNameFromDst(funcDecl *dst.FuncDecl) string {
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		if ident, ok := funcDecl.Recv.List[0].Type.(*dst.Ident); ok {
			return ident.Name + "." + funcDecl.Name.Name
		}
	}
	return funcDecl.Name.Name
}

func funcNameFromTypes(funcObj *types.Func) string {
	sig, ok := funcObj.Type().(*types.Signature)
	if !ok || sig.Recv() == nil {
		return funcObj.Name()
	}
	t := sig.Recv().Type()
	switch receiver := t.(type) {
	case *types.Named:
		return fmt.Sprintf("%s.%s", receiver.Obj().Name(), funcObj.Name())
	case *types.Pointer:
		var named = receiver.Elem().(*types.Named)
		return fmt.Sprintf("*%s.%s", named.Obj().Name(), funcObj.Name())
	default:
		panic(fmt.Sprintf("unhandled receiver type: %T", receiver))
	}

}

func (r *Registry) scanFile(file *dst.File, pkg *decorator.Package) (result map[TypeID]*FuncEntry, err error) {
	var (
		importMap = NewImportMap(pkg.PkgPath, file.Imports)
	)
	result = make(map[TypeID]*FuncEntry)

	for _, _decl := range file.Decls {
		switch decl := _decl.(type) {
		case *dst.FuncDecl:
			if isRelevantFunc(decl) {
				typeID := NewTypeID(pkg.PkgPath, funcNameFromDst(decl))
				result[typeID] = &FuncEntry{ImportMap: importMap, FuncDecl: decl, typeID: typeID}
			}
		case *dst.GenDecl:
			switch decl.Tok {
			case token.CONST:
				if err := r.registerConstDecl(file, pkg, decl); err != nil {
					return result, fmt.Errorf("register const: %w", err)
				}
			case token.TYPE:
				if err := r.registerTypeDecl(file, pkg, decl); err != nil {
					return result, fmt.Errorf("registering type decl %v in file %s (Pkg %s): %w", decl, file.Name.Name, pkg.PkgPath, err)
				}
			case token.VAR:
				if err := r.registerVarDecl(file, pkg, decl, importMap); err != nil {
					return result, err
				}
			default:
				continue
			}
		}

	}

	return result, nil
}

func (r *Registry) registerVarDecl(file *dst.File, pkg *decorator.Package, decl *dst.GenDecl, importMap ImportMap) error {
	for _, spec := range decl.Specs {
		valueSpec := spec.(*dst.ValueSpec)
		for _, val := range valueSpec.Values {
			if callExpr, ok := val.(*dst.CallExpr); ok && isUnionTypeDecl(callExpr, importMap) {
				// nodeImpl has been identified as a Union Type declaration.  Note the arguments.
				if err := r.registerUnionTypeDecl(file, pkg, callExpr, importMap); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// isUnionTypeDecl determines whether the *dst.CallExpr is a call to
// jsonschema.SetTypeAlternative, by checking the selector expression.
func isUnionTypeDecl(callExpr *dst.CallExpr, importMap ImportMap) bool {
	indexExpr, ok := callExpr.Fun.(*dst.IndexExpr)
	if !ok {
		return false
	}
	ident, ok := indexExpr.X.(*dst.Ident)
	if !ok {
		return false
	}
	return ident.Name == UnionTypeFunc && ident.Path == JSONSchemaPackage
}

func (r *Registry) registerUnionTypeDecl(file *dst.File, pkg *decorator.Package, callExpr *dst.CallExpr, importMap ImportMap) error {
	indexExpr, ok := callExpr.Fun.(*dst.IndexExpr)
	if !ok {
		panic("that should not be")
	}
	unionTypeDecl := SetTypeAlternativeDecl(importMap, indexExpr.Index)
	for _, arg := range callExpr.Args {
		alt, err := r.interpretUnionTypeAltArg(arg, importMap)
		if err != nil {
			return err
		}
		unionTypeDecl.Alternatives = append(unionTypeDecl.Alternatives, alt)
	}
	unionTypeDecl.File = file
	unionTypeDecl.Pkg = pkg
	r.unionTypes[unionTypeDecl.ID()] = unionTypeDecl

	return nil
}

var ErrInvalidUnionTypeArg = errors.New("invalid union type arg")

func (r *Registry) interpretUnionTypeAltArg(expr dst.Expr, importMap ImportMap) (alt TypeAlternative, err error) {
	callExpr, ok := expr.(*dst.CallExpr)
	if !ok {
		return alt, ErrInvalidUnionTypeArg
	}
	switch fun := callExpr.Fun.(type) {
	case *dst.Ident:
		if fun.Name != TypeAltFunc || fun.Path != JSONSchemaPackage {
			return alt, ErrInvalidUnionTypeArg
		}
	default:
		return alt, ErrInvalidUnionTypeArg
	}
	alt.Alias = callExpr.Args[0].(*dst.BasicLit).Value
	alt.Alias = alt.Alias[1 : len(alt.Alias)-1]
	alt.ImportMap = importMap

	switch typeArg := callExpr.Args[1].(type) {
	case *dst.SelectorExpr:
		// This is the case of a struct method whose receiver is the
		// alternate type.
		if ident, ok := typeArg.X.(*dst.Ident); ok {
			alt.FuncName = fmt.Sprintf("%s.%s", ident.Name, typeArg.Sel.Name)
		} else {
			return alt, ErrInvalidUnionTypeArg
		}
	case *dst.Ident:
		alt.FuncName = typeArg.Name
	default:
		return alt, ErrInvalidUnionTypeArg
	}

	return alt, nil
}

func (r *Registry) registerTypeDecl(file *dst.File, pkg *decorator.Package, genDecl *dst.GenDecl) error {
	for _, spec := range toTypeSpecs(genDecl.Specs) {
		//inspect("Scanned spec", spec)
		ts := &typeSpec{
			typeSpec: spec,
			pkg:      pkg,
			genDecl:  genDecl,
			file:     file,
		}
		r.typeMap[ts.ID()] = ts
	}
	return nil
}

func toTypeSpecs(specs []dst.Spec) (ts []*dst.TypeSpec) {
	for _, spec := range specs {
		ts = append(ts, spec.(*dst.TypeSpec))
	}
	return ts
}

func inspect(str string, item ...any) {
	fmt.Printf("inspect %s: %T %v\n", str, item[0], item[0])
	for i := 1; i < len(item); i++ {
		fmt.Printf("     item[%d]: %T %v\n", i, item[i], item[i])
	}
}
