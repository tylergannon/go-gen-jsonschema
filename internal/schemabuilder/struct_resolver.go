package schemabuilder

import (
	"errors"
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"go/token"
	"golang.org/x/tools/go/packages"
	"slices"
	"strings"
)

// resolveTypeSpec finds the TypeSpec node for a given type name in the package.
func resolveTypeSpec(pkg *decorator.Package, typeName string) *dst.TypeSpec {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*dst.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			// If there is exactly one spec, and it doesn't have its own
			// Start decoration, transfer the Start decoration from the
			// Gen Decl to the TypeSpec.
			idx := slices.IndexFunc(genDecl.Specs, makeTypeSpecFinder(typeName))
			if idx == -1 {
				continue
			}
			typeSpec := genDecl.Specs[idx].(*dst.TypeSpec)
			decs := typeSpec.Decorations()
			if len(genDecl.Specs) == 1 && len(decs.Start) == 0 {
				decs.Start.Replace(genDecl.Decs.Start...)
			}
			return typeSpec
		}
	}
	return nil
}

func makeTypeSpecFinder(name string) func(it dst.Spec) bool {
	return func(it dst.Spec) bool {
		if typeSpec, ok := it.(*dst.TypeSpec); ok {
			return typeSpec.Name.Name == name
		}
		return false
	}
}

type StructTypeData struct {
	StructType *dst.StructType
	TypeSpec   *dst.TypeSpec
}

// resolveStruct recursively resolves a struct type, handling definitions and aliases.
func resolveStruct(importMap *PackageMap, typeName string, path string) (result *StructTypeData, err error) {
	var (
		pkg = importMap.GetPackage(path)
	)
	if pkg == nil {
		return nil, fmt.Errorf("no package loaded for path %s", path)
	}
	typeSpec := resolveTypeSpec(pkg, typeName)
	if typeSpec == nil {
		return nil, fmt.Errorf("TypeSpec for %s not found", typeName)
	}
	result = &StructTypeData{TypeSpec: typeSpec}

	switch t := typeSpec.Type.(type) {
	case *dst.StructType:
		// Base case: Found the struct definition
		result.StructType = t
	case *dst.Ident:
		// Recursive case: Follow the referenced type

		if t.Obj != nil {
			result, err = resolveStruct(importMap, t.Name, path)
		} else if t.Path != "" {
			result, err = resolveStruct(importMap, t.Name, t.Path)
		} else {
			err = errors.New("not sure what happened")
		}
		if err != nil {
			return nil, err
		}
	case *dst.SelectorExpr:
		// Handle types defined in other packages (subpkg.Type)
		// NOTE: This assumes that you've already loaded the relevant package's AST.
		pkgName := t.X.(*dst.Ident).Name
		subPkg := getImportedPackage(importMap, pkgName)
		if subPkg == nil {
			return nil, fmt.Errorf("Imported package %s not found", pkgName)
		}
		result, err = resolveStruct(importMap, t.Sel.Name, path)
	default:
		return nil, fmt.Errorf("Unsupported type for %s", typeName)
	}
	result.TypeSpec = typeSpec
	return result, nil
}

// getImportedPackage resolves an imported package from the current AST.
func getImportedPackage(importMap *PackageMap, importName string) *packages.Package {
	// Example: look through imports and resolve the associated package.
	// This function needs the imports loaded and parsed to work correctly.
	// Stubbed here for simplicity.
	return nil
}

// extractFieldComments collects field comments from a StructType.
func extractFieldComments(structType *dst.StructType) map[string]string {
	fieldComments := make(map[string]string)
	for _, field := range structType.Fields.List {
		if len(field.Names) > 0 {
			var comments []string
			fieldName := field.Names[0].Name
			if field.Decorations().Start != nil {
				comments = append(comments, field.Decorations().Start.All()...)
			}
			if field.Decorations().End != nil {
				comments = append(comments, field.Decorations().End.All()...)
			}
			for i := 0; i < len(comments); i++ {
				comments[i] = strings.TrimPrefix(comments[i], "// ")
				comments[i] = strings.TrimPrefix(comments[i], "//")
			}
			if len(comments) > 0 {
				fieldComments[fieldName] = strings.Join(comments, "\n")
			} else {
				fieldComments[fieldName] = ""
			}
		}
	}
	return fieldComments
}
