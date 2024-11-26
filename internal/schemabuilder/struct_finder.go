package schemabuilder

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"log"
	"strings"
)

// typeInfo holds the type and its associated documentation
type typeInfo struct {
	named    *types.Named
	comments string
}

// findTypeInfoInPackage looks for a type declaration and its associated comments
func findTypeInfoInPackage(pkg *packages.Package, typeName string) (*typeInfo, error) {
	var result *typeInfo

	// First find the named type
	named, err := findNamedType(pkg, typeName)
	if err != nil {
		return nil, err
	}
	log.Println(named)

	// Now search through all syntax files for the type declaration and comments
	for _, file := range pkg.Syntax {
		result = findStructDeclInFile(file, named)
		if result != nil {
			return result, nil
		}
	}

	return nil, fmt.Errorf("could not find type %q in AST", typeName)
}
func findStructDeclInFile(file *ast.File, named *types.Named) *typeInfo {
	var (
		comments []string
		result   *typeInfo
	)

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		// Collect comments
		if comment, ok := n.(*ast.Comment); ok {
			comments = append(comments, strings.TrimPrefix(comment.Text, "// "))
			return true
		}

		if genDecl, ok := n.(*ast.GenDecl); ok {
			if genDecl.Tok != token.TYPE {
				return true
			}
			for _, spec := range genDecl.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					_ = ts

				}
			}
		}

		// Check for our type
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if typeSpec.Name.Name == named.Obj().Name() {
				result = &typeInfo{
					named:    named,
					comments: strings.Join(comments, "\n"),
				}
				return false // Stop traversal
			}
		}

		// Clear comments after any non-comment node
		comments = nil
		return true
	})
	return result
}
