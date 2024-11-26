package schemabuilder

import (
	"fmt"
	"go/ast"
	"slices"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/tools/go/packages"
)

// resolveTypeSpec finds the TypeSpec node for a given type name in the package.
func resolveTypeSpec(pkg *packages.Package, typeName string) *ast.TypeSpec {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if ok && typeSpec.Name.Name == typeName {
					return typeSpec
				}
			}
		}
	}
	return nil
}

// resolveStruct recursively resolves a struct type, handling definitions and aliases.
func resolveStruct(importMap *ImportMap, typeName string) (*ast.StructType, error) {
	typeSpec := resolveTypeSpec(importMap.localPackage, typeName)
	if typeSpec == nil {
		return nil, fmt.Errorf("TypeSpec for %s not found", typeName)
	}

	switch t := typeSpec.Type.(type) {
	case *ast.StructType:
		// Base case: Found the struct definition
		return t, nil
	case *ast.Ident:
		// Recursive case: Follow the referenced type
		return resolveStruct(importMap, t.Name)
	case *ast.SelectorExpr:
		// Handle types defined in other packages (subpkg.Type)
		// NOTE: This assumes that you've already loaded the relevant package's AST.
		pkgName := t.X.(*ast.Ident).Name
		subPkg := getImportedPackage(pkg, pkgName)
		if subPkg == nil {
			return nil, fmt.Errorf("Imported package %s not found", pkgName)
		}
		return resolveStruct(importMap, t.Sel.Name)
	default:
		return nil, fmt.Errorf("Unsupported type for %s", typeName)
	}
}

// getImportedPackage resolves an imported package from the current AST.
func getImportedPackage(importMap, importName string) *packages.Package {
	// Example: look through imports and resolve the associated package.
	// This function needs the imports loaded and parsed to work correctly.
	// Stubbed here for simplicity.
	return nil
}

// extractFieldComments collects field comments from a StructType.
func extractFieldComments(structType *ast.StructType) map[string]string {
	fieldComments := make(map[string]string)
	for _, field := range structType.Fields.List {
		if len(field.Names) > 0 {
			fieldName := field.Names[0].Name
			if field.Comment != nil {
				fieldComments[fieldName] = strings.TrimSpace(field.Comment.Text())
			} else if field.Doc != nil {
				fieldComments[fieldName] = strings.TrimSpace(field.Doc.Text())
			} else {
				fieldComments[fieldName] = ""
			}
		}
	}
	return fieldComments
}

var _ = Describe("AST Struct Resolution", func() {
	var (
		importMap *ImportMap
	)
	BeforeEach(func() {
		var err error
		// Load the package with its AST
		pkgs, err := packages.Load(DefaultPackageCfg, "./fixtures/testapp1/...")
		Expect(err).To(BeNil())
		Expect(pkgs).ToNot(BeEmpty())
		mainPkgIdx := slices.IndexFunc(pkgs, func(p *packages.Package) bool {
			return p.Name == "testapp1"
		})
		Expect(mainPkgIdx).To(BeNumerically(">=", 0))
		importMap = NewImportMap(pkgs[mainPkgIdx])
		for _, pkg := range pkgs {
			importMap.AddPackage(pkg)
		}
	})
	DescribeTable("should resolve struct definitions and field comments", func(fieldName string) {
		structType, err := resolveStruct(mainPkg, fieldName)
		Expect(err).To(BeNil())

		fieldComments := extractFieldComments(structType)
		Expect(fieldComments).To(HaveKeyWithValue("Foo", "There can be comments here"))
		Expect(fieldComments).To(HaveKeyWithValue("Bar", "There can also be comments to the right"))
		Expect(fieldComments).To(HaveKeyWithValue("Baz", "But in that case, this will be ignored."))
		//Expect(fieldComments).To(HaveKeyWithValue("DefinedElsewhere", ""))
		Expect(fieldComments).To(HaveKey("DefinedElsewhere"))
	},
		Entry("Actual Definition", "ComplexExample"),
		Entry("Type Definition", "ComplexDefinition"),
		Entry("Type Alias", "ComplexAlias"),
	)
	DescribeTable("resolving struct definitions from another package", func(fieldName string) {
		_, err := resolveStruct(mainPkg, fieldName)
		Expect(err).To(BeNil())

		//fieldComments := extractFieldComments(structType)
		//Expect(fieldComments).To(HaveKeyWithValue("Foo", "There can be comments here"))
		//Expect(fieldComments).To(HaveKeyWithValue("Bar", "There can also be comments to the right"))
		//Expect(fieldComments).To(HaveKeyWithValue("Baz", "But in that case, this will be ignored."))
		////Expect(fieldComments).To(HaveKeyWithValue("DefinedElsewhere", ""))
		//Expect(fieldComments).To(HaveKey("DefinedElsewhere"))
	},
		Entry("Remote Definition", "RemoteDefinition"),
		//Entry("Type Definition", "ComplexDefinition"),
		//Entry("Type Alias", "ComplexAlias"),
	)
})
