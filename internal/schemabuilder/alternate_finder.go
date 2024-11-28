package schemabuilder

import (
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"go/token"
	"strings"
)

// AltFunction represents a single alternative in a union type definition
type AltFunction struct {
	Name          string // The name passed to Alt()
	ConverterType string // The type containing the conversion method (e.g., "TimeAgo")
	ConverterPkg  string // Package of the converter type, if external
	ConverterFunc string // The method name (e.g., "ToTime")
}

// UnionTypeInfo represents all alternatives for a union type
type UnionTypeInfo struct {
	TypeName     string
	Alternatives []AltFunction
}

// FindUnionType analyzes a package for union type definitions
func FindUnionType(typeName string, pkg *decorator.Package) (*UnionTypeInfo, error) {
	result := &UnionTypeInfo{
		TypeName:     typeName,
		Alternatives: make([]AltFunction, 0),
	}

	// Walk through all files in the package
	for _, file := range pkg.Syntax {
		imports := NewImportMap(pkg.PkgPath, file.Imports)

		dst.Inspect(file, func(n dst.Node) bool {
			// Look for variable declarations (var _ = jsonschema.NewUnionType...)
			if assign, ok := n.(*dst.AssignStmt); ok {
				if len(assign.Rhs) != 1 {
					return true
				}

				// Check if it's a call to NewUnionType
				call, ok := assign.Rhs[0].(*dst.CallExpr)
				if !ok {
					return true
				}

				// Verify it's jsonschema.NewUnionType
				if !isNewUnionTypeCall(call) {
					return true
				}

				// Extract type parameter to verify it matches our target type
				if !matchesTargetType(call, typeName) {
					return true
				}

				// Process each Alt() argument
				for _, arg := range call.Args {
					altCall, ok := arg.(*dst.CallExpr)
					if !ok {
						continue
					}

					alt := processAltCall(altCall, imports)
					if alt != nil {
						result.Alternatives = append(result.Alternatives, *alt)
					}
				}
			}
			return true
		})
	}

	if len(result.Alternatives) == 0 {
		return nil, fmt.Errorf("no union type definition found for type %s", typeName)
	}

	return result, nil
}

func isNewUnionTypeCall(call *dst.CallExpr) bool {
	if sel, ok := call.Fun.(*dst.SelectorExpr); ok {
		if ident, ok := sel.X.(*dst.Ident); ok {
			return ident.Name == "jsonschema" && sel.Sel.Name == "NewUnionType"
		}
	}
	return false
}

func matchesTargetType(call *dst.CallExpr, targetType string) bool {
	// This would need to analyze the type parameter of NewUnionType
	// For simplicity, we're assuming the union type definition we find is for our target
	// A more robust implementation would verify the type parameter
	return true
}

func processAltCall(call *dst.CallExpr, imports ImportMap) *AltFunction {
	if sel, ok := call.Fun.(*dst.SelectorExpr); ok {
		if ident, ok := sel.X.(*dst.Ident); ok {
			if ident.Name == "jsonschema" && sel.Sel.Name == "Alt" {
				if len(call.Args) != 2 {
					return nil
				}

				// Extract the name string literal
				nameArg, ok := call.Args[0].(*dst.BasicLit)
				if !ok || nameArg.Kind != token.STRING {
					return nil
				}
				name := strings.Trim(nameArg.Value, "\"")

				// Process the function reference
				funcArg := call.Args[1]
				converter := extractConverterInfo(funcArg)

				if converter != nil {
					return &AltFunction{
						Name:          name,
						ConverterType: converter.Type,
						ConverterPkg:  imports[converter.Package],
						ConverterFunc: converter.Method,
					}
				}
			}
		}
	}
	return nil
}

type converterInfo struct {
	Package string
	Type    string
	Method  string
}

func extractConverterInfo(expr dst.Expr) *converterInfo {
	// Handle selector expressions like TimeAgo.ToTime or pkg.Type.Method
	if sel, ok := expr.(*dst.SelectorExpr); ok {
		if typeIdent, ok := sel.X.(*dst.Ident); ok {
			// Simple case: Type.Method
			return &converterInfo{
				Type:   typeIdent.Name,
				Method: sel.Sel.Name,
			}
		} else if pkgSel, ok := sel.X.(*dst.SelectorExpr); ok {
			// Package case: pkg.Type.Method
			if pkgIdent, ok := pkgSel.X.(*dst.Ident); ok {
				return &converterInfo{
					Package: pkgIdent.Name,
					Type:    pkgSel.Sel.Name,
					Method:  sel.Sel.Name,
				}
			}
		}
	}
	return nil
}
