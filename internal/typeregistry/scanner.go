package typeregistry

import (
	"fmt"
	"go/token"
	"go/types"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/tylergannon/go-gen-jsonschema/internal/loader"
)

const (
	JSONSchemaPackage = "github.com/tylergannon/go-gen-jsonschema"
	// UnionTypeFunc Used for denoting basic alts for a struct type.  These are special because each "type alt"
	// needs to have a converter func that renders it into the destination type.
	UnionTypeFunc = "SetTypeAlternative"
	// ImplementationFunc denotes the list of acceptable implementations for an interface type.
	// These are different from the "type alt" in that there is no conversion function.  By being
	// an implementation of the denoted interface, they are directly assignable to that type.
	SetImplementationFunc = "SetImplementations"
	TypeAltFunc           = "Alt"
)

func NewRegistry(pkgs []*decorator.Package) (*Registry, error) {
	registry := &Registry{
		typeMap:        map[TypeID]*typeSpec{},
		packages:       map[string]*decorator.Package{},
		unionTypes:     map[TypeID]*UnionTypeDecl{},
		interfaceTypes: map[TypeID]*InterfaceTypeDecl{},
		imports:        map[string]*decorator.Package{},
		constants:      map[TypeID][]EnumEntry{},
		funcs:          map[TypeID]*FuncEntry{},
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
	funcTypes := findAllFuncsInPkg(pkg)

	for _, file := range pkg.Syntax {
		if funcsTemp, err := r.scanFile(file, funcTypes, pkg); err != nil {
			return err
		} else {
			for id, funcDecl := range funcsTemp {
				r.funcs[id] = funcDecl
			}
		}
	}

	r.packages[pkg.PkgPath] = pkg

	return nil
}

func findAllFuncsInPkg(pkg *decorator.Package) map[TypeID]*types.Func {
	funcs := map[TypeID]*types.Func{}

	for _, obj := range pkg.TypesInfo.Defs {
		if obj == nil {
			continue
		}

		funcObj, ok := obj.(*types.Func)
		if !ok {
			continue
		}
		typeID := NewTypeID(pkg.PkgPath, funcNameFromTypes(funcObj))
		funcs[typeID] = funcObj
	}
	return funcs
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
		named := receiver.Elem().(*types.Named)
		return fmt.Sprintf("*%s.%s", named.Obj().Name(), funcObj.Name())
	default:
		panic(fmt.Sprintf("unhandled receiver type: %T", receiver))
	}
}

func (r *Registry) scanFile(file *dst.File, funcTypes map[TypeID]*types.Func, pkg *decorator.Package) (result map[TypeID]*FuncEntry, err error) {
	importMap := NewImportMap(pkg.PkgPath, file.Imports)
	result = make(map[TypeID]*FuncEntry)

	for _, _decl := range file.Decls {
		switch decl := _decl.(type) {
		case *dst.FuncDecl:
			var entry = NewFuncEntry(decl, pkg, file, importMap)
			if typesFunc, ok := funcTypes[entry.typeID]; ok {
				entry.Func = typesFunc
			} else {
				continue // not sure why this would ever happen; just keeping the logic the same for now.
			}
			if entry.isCandidateAltConverter() {
				result[entry.typeID] = entry
			} else if typeName, pkgPath, ok := entry.IsUnmarshalJSON(); ok {
				r.unmarshalers[NewTypeID(pkgPath, typeName)] = entry
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

func (r *Registry) registerVarDecl(file *dst.File, pkg *decorator.Package, decl *dst.GenDecl, importMap ImportMap) (err error) {
	for _, spec := range decl.Specs {
		valueSpec := spec.(*dst.ValueSpec)
		for _, val := range valueSpec.Values {
			callExpr, ok := val.(*dst.CallExpr)
			if !ok {
				continue
			}
			var funcName string
			if funcName, ok = isUnionTypeDecl(callExpr, importMap); !ok {
				continue
			}
			switch funcName {
			case UnionTypeFunc:
				err = r.registerUnionTypeDecl(file, pkg, callExpr, importMap)
			case SetImplementationFunc:
				err = r.registerInterfaceDeclaration(file, pkg, callExpr, importMap)
			}
			// nodeImpl has been identified as a Union Type declaration.  Note the arguments.
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// isUnionTypeDecl determines whether the *dst.CallExpr is a call to
// jsonschema.SetTypeAlternative, by checking the selector expression.
func isUnionTypeDecl(callExpr *dst.CallExpr, importMap ImportMap) (funcName string, isUnionTypeDecl bool) {
	indexExpr, ok := callExpr.Fun.(*dst.IndexExpr)
	if !ok {
		return "", false
	}
	ident, ok := indexExpr.X.(*dst.Ident)
	if !ok {
		return "", false
	}
	if (ident.Name == UnionTypeFunc || ident.Name == SetImplementationFunc) && ident.Path == JSONSchemaPackage {
		return ident.Name, true
	}
	return "", false
}

// registerUnionTypeDecl is for registering a union type that converts to a
// struct by means of conversion functions.
func (r *Registry) registerUnionTypeDecl(file *dst.File, pkg *decorator.Package, callExpr *dst.CallExpr, importMap ImportMap) (err error) {
	indexExpr, ok := callExpr.Fun.(*dst.IndexExpr)
	if !ok {
		panic("that should not be")
	}

	unionTypeDecl := SetTypeAlternativeDecl(importMap, indexExpr)
	for _, arg := range callExpr.Args {
		var alt TypeAlternative
		if alt, err = r.interpretUnionTypeAltArg(arg, importMap); err != nil {
			return err
		}
		unionTypeDecl.Alternatives = append(unionTypeDecl.Alternatives, alt)
	}
	unionTypeDecl.File = file
	unionTypeDecl.Pkg = pkg
	r.unionTypes[unionTypeDecl.ID()] = unionTypeDecl

	return nil
}

func (r *Registry) registerTypeDecl(file *dst.File, pkg *decorator.Package, genDecl *dst.GenDecl) error {
	for _, spec := range toTypeSpecs(genDecl.Specs) {
		// inspect("Scanned spec", spec)
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
	fmt.Printf("\ninspect %s: %T %v\n", str, item[0], item[0])
	for i := 1; i < len(item); i++ {
		fmt.Printf("     item[%d]: %T %v\n", i, item[i], item[i])
	}
}
