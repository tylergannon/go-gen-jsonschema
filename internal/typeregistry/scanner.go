package typeregistry

import (
	"errors"
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/tylergannon/go-gen-jsonschema/internal/loader"
	"go/token"
	"log"
)

const (
	JSONSchemaPackage = "github.com/tylergannon/go-gen-jsonschema"
	UnionTypeFunc     = "SetTypeAlternative"
	TypeAltFunc       = "Alt"
)

func NewRegistry(pkgs []*decorator.Package) (*Registry, error) {
	registry := &Registry{
		typeMap:    map[TypeID]TypeSpec{},
		packages:   map[string]*decorator.Package{},
		unionTypes: map[TypeID]*UnionTypeDecl{},
		imports:    map[string]*decorator.Package{},
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
	for _, file := range pkg.Syntax {
		if err := r.scanFile(file, pkg); err != nil {
			return err
		}
	}
	r.packages[pkg.PkgPath] = pkg
	for _, pkg := range pkg.Imports {
		r.imports[pkg.PkgPath] = pkg
	}
	return nil
}

func (r *Registry) scanFile(file *dst.File, pkg *decorator.Package) error {
	importMap := NewImportMap(pkg.PkgPath, file.Imports)
	for _, decl := range getTypeDecls(file) {
		switch decl.Tok {
		case token.TYPE:
			if err := r.registerTypeDecl(file, pkg, decl); err != nil {
				return fmt.Errorf("registering type decl %v in file %s (pkg %s): %w", decl, file.Name.Name, pkg.PkgPath, err)
			}
		case token.VAR:
			if err := r.registerVarDecl(file, pkg, decl, importMap); err != nil {
				return err
			}
		default:
			panic("no implementation for token type " + decl.Tok.String())
		}
	}
	return nil
}

func (r *Registry) registerVarDecl(file *dst.File, pkg *decorator.Package, decl *dst.GenDecl, importMap ImportMap) error {
	for _, spec := range decl.Specs {
		valueSpec := spec.(*dst.ValueSpec)
		for _, val := range valueSpec.Values {
			if callExpr, ok := val.(*dst.CallExpr); ok && isUnionTypeDecl(callExpr, importMap) {
				// Node has been identified as a Union Type declaration.  Note the arguments.
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
		alt, err := interpretUnionTypeAltArg(arg, importMap)
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

func interpretUnionTypeAltArg(expr dst.Expr, importMap ImportMap) (alt TypeAlternative, err error) {
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

	switch typeArg := callExpr.Args[1].(type) {
	case *dst.SelectorExpr:
		if ident, ok := typeArg.X.(*dst.Ident); ok {
			alt.PkgPath = ident.Path
			if alt.PkgPath == "" {
				alt.PkgPath = importMap[""]
			}
			alt.TypeName = ident.Name
			alt.FuncName = typeArg.Sel.Name
		} else {
			return alt, ErrInvalidUnionTypeArg
		}
	case *dst.Ident:
		alt.PkgPath = importMap[""]
		alt.FuncName = typeArg.Name
	default:
		return alt, ErrInvalidUnionTypeArg
	}

	return alt, nil
}

func (r *Registry) registerTypeDecl(file *dst.File, pkg *decorator.Package, genDecl *dst.GenDecl) error {
	for _, spec := range toTypeSpecs(genDecl.Specs) {
		inspect("Scanned spec", spec)
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

// getTypeDecls locates all GenDecl nodes for TYPE and VAR declarations
func getTypeDecls(file *dst.File) (decls []*dst.GenDecl) {
	var (
		genDecl *dst.GenDecl
		ok      bool
	)
	for _, decl := range file.Decls {
		if genDecl, ok = decl.(*dst.GenDecl); !ok {
			continue
		}
		if genDecl.Tok == token.TYPE || genDecl.Tok == token.VAR {
			decls = append(decls, genDecl)
		}
	}
	return decls
}

func toTypeSpecs(specs []dst.Spec) (ts []*dst.TypeSpec) {
	for _, spec := range specs {
		ts = append(ts, spec.(*dst.TypeSpec))
	}
	return ts
}

func inspect(str string, item any) {
	log.Printf("inspect %s: %T %v\n", str, item, item)
}
