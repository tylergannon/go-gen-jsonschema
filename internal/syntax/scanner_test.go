package syntax

import (
	"fmt"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/stretchr/testify/require"
)

const (
	pkgPath = "github.com/tylergannon/go-gen-jsonschema/internal/syntax/testfixtures/typescanner"
	subpkg  = "github.com/tylergannon/go-gen-jsonschema/internal/syntax/testfixtures/typescanner/scannersubpkg"
)

type fileSpecs struct {
	file    *dst.File
	specs   []dst.Spec
	genDecl *dst.GenDecl
	pkg     *decorator.Package
}

func LoadDecls(path, fileName string, tok token.Token) []fileSpecs {
	var result []fileSpecs
	pkgs, err := Load(path)
	if err != nil {
		panic(err)
	}
	for _, pkg := range pkgs {
		_, err = LoadPackage(pkg)
		if err != nil {
			panic(err)
		}
		for _, file := range pkg.Syntax {
			pos := nodePosition(pkg, file)
			if filepath.Base(pos.Filename) != fileName {
				continue
			}
			var fileSpec fileSpecs
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*dst.GenDecl)
				if !ok || genDecl.Tok != tok {
					continue
				}
				fileSpec.genDecl = genDecl
				fileSpec.specs = append(fileSpec.specs, genDecl.Specs...)
			}
			if len(fileSpec.specs) > 0 {
				fileSpec.file = file
				fileSpec.pkg = pkg
				result = append(result, fileSpec)
			}
		}
	}
	return result
}

func TestFuncCallParser(t *testing.T) {
	specs := LoadDecls("./testfixtures/typescanner", "calls.go", token.VAR)
	require.Len(t, specs, 1)
	require.Len(t, specs[0].specs, 10)
	valueSpec := func(idx int) ValueSpec {
		return NewValueSpec(specs[0].genDecl, specs[0].specs[idx].(*dst.ValueSpec), specs[0].pkg, specs[0].file)
	}

	var calls []MarkerFunctionCall
	for _, spec := range specs[0].specs {
		calls = append(calls, ParseValueExprForMarkerFunctionCall(NewValueSpec(specs[0].genDecl, spec.(*dst.ValueSpec), specs[0].pkg, specs[0].file))...)
	}
	require.Len(t, calls, 10)

	t.Run("Call number 1", func(t *testing.T) {
		call := ParseValueExprForMarkerFunctionCall(valueSpec(0))[0]
		require.Equal(t, MarkerFuncNewJSONSchemaMethod, call.CallExpr.MustIdentifyFunc().TypeName)
		require.Len(t, call.CallExpr.Args(), 1)
		require.Nil(t, call.TypeArgument())
		schemaMethod, err := call.ParseSchemaMethod()
		require.NoError(t, err)
		require.Equal(t, "Schema", schemaMethod.SchemaMethodName)
		require.Equal(t, NormalConcrete, schemaMethod.Receiver.Indirection)
		require.Equal(t, pkgPath, schemaMethod.Receiver.PkgPath)
		require.Equal(t, "TypeForSchemaMethod", schemaMethod.Receiver.TypeName)
	})

	t.Run("Call number 2", func(t *testing.T) {
		call := ParseValueExprForMarkerFunctionCall(valueSpec(1))[0]
		require.Equal(t, MarkerFuncNewJSONSchemaMethod, call.CallExpr.MustIdentifyFunc().TypeName)
		require.Len(t, call.CallExpr.Args(), 1)
		require.Nil(t, call.TypeArgument())
		schemaMethod, err := call.ParseSchemaMethod()
		require.NoError(t, err)
		require.Equal(t, "Schema", schemaMethod.SchemaMethodName)
		require.Equal(t, Pointer, schemaMethod.Receiver.Indirection)
		require.Equal(t, pkgPath, schemaMethod.Receiver.PkgPath)
		require.Equal(t, "PointerTypeForSchemaMethod", schemaMethod.Receiver.TypeName)
	})

	t.Run("Call number 3", func(t *testing.T) {
		call := ParseValueExprForMarkerFunctionCall(valueSpec(2))[0]
		require.Equal(t, MarkerFuncNewJSONSchemaBuilder, call.CallExpr.MustIdentifyFunc().TypeName)
		require.Len(t, call.CallExpr.Args(), 1)
		require.NotNil(t, call.MustTypeArgument())
		require.Equal(t, "TypeForSchemaFunction", call.MustTypeArgument().TypeName)
		require.Equal(t, pkgPath, call.MustTypeArgument().PkgPath)
	})

	t.Run("Call number 4", func(t *testing.T) {
		call := ParseValueExprForMarkerFunctionCall(valueSpec(3))[0]
		require.Equal(t, MarkerFuncNewJSONSchemaBuilder, call.CallExpr.MustIdentifyFunc().TypeName)
		require.Len(t, call.CallExpr.Args(), 1)
		require.NotNil(t, call.MustTypeArgument())
		require.Equal(t, "PointerTypeForSchemaFunction", call.MustTypeArgument().TypeName)
		require.Equal(t, Pointer, call.MustTypeArgument().Indirection)
		require.Equal(t, pkgPath, call.MustTypeArgument().PkgPath)
	})

	t.Run("Call number 5", func(t *testing.T) {
		call := ParseValueExprForMarkerFunctionCall(valueSpec(4))[0]
		require.Equal(t, MarkerFuncNewInterfaceImpl, call.CallExpr.MustIdentifyFunc().TypeName)
		require.Len(t, call.CallExpr.Args(), 4)
		require.NotNil(t, call.MustTypeArgument())
		require.Equal(t, "MarkerInterface", call.MustTypeArgument().TypeName)
		require.Equal(t, NormalConcrete, call.MustTypeArgument().Indirection)
		require.Equal(t, pkgPath, call.MustTypeArgument().PkgPath)
		callArgs, err := call.ParseTypesFromArgs()
		require.NoError(t, err)
		require.NotEmpty(t, callArgs)
		require.Equal(t, "Type001", callArgs[0].TypeName)
		require.Equal(t, "Type002", callArgs[1].TypeName)
		require.Equal(t, "Type003", callArgs[2].TypeName)
		require.Equal(t, "Type004", callArgs[3].TypeName)
	})

	t.Run("Call number 6", func(t *testing.T) {
		call := ParseValueExprForMarkerFunctionCall(valueSpec(5))[0]
		require.Equal(t, MarkerFuncNewEnumType, call.CallExpr.MustIdentifyFunc().TypeName)
		require.Len(t, call.CallExpr.Args(), 0)
		require.NotNil(t, call.MustTypeArgument())
		require.Equal(t, "NiceEnumType", call.MustTypeArgument().TypeName)
		require.Equal(t, NormalConcrete, call.MustTypeArgument().Indirection)
		require.Equal(t, pkgPath, call.MustTypeArgument().PkgPath)
	})

	t.Run("Call number 7", func(t *testing.T) {
		call := ParseValueExprForMarkerFunctionCall(valueSpec(6))[0]
		require.Equal(t, MarkerFuncNewJSONSchemaBuilder, call.CallExpr.MustIdentifyFunc().TypeName)
		require.Len(t, call.CallExpr.Args(), 1)
		require.NotNil(t, call.MustTypeArgument())
		require.Equal(t, "TypeForSchemaFunction", call.MustTypeArgument().TypeName)
		require.Equal(t, NormalConcrete, call.MustTypeArgument().Indirection)
		require.Equal(t, subpkg, call.MustTypeArgument().PkgPath)
	})

	t.Run("Call number 8", func(t *testing.T) {
		call := ParseValueExprForMarkerFunctionCall(valueSpec(7))[0]
		require.Equal(t, MarkerFuncNewJSONSchemaBuilder, call.CallExpr.MustIdentifyFunc().TypeName)
		require.Len(t, call.CallExpr.Args(), 1)
		require.NotNil(t, call.MustTypeArgument())
		require.Equal(t, "PointerTypeForSchemaFunction", call.MustTypeArgument().TypeName)
		require.Equal(t, Pointer, call.MustTypeArgument().Indirection)
		require.Equal(t, subpkg, call.MustTypeArgument().PkgPath)
	})

	t.Run("Call number 9", func(t *testing.T) {
		call := ParseValueExprForMarkerFunctionCall(valueSpec(8))[0]
		require.Equal(t, MarkerFuncNewInterfaceImpl, call.CallExpr.MustIdentifyFunc().TypeName)
		require.Len(t, call.CallExpr.Args(), 4)
		require.NotNil(t, call.MustTypeArgument())
		require.Equal(t, "MarkerInterface", call.MustTypeArgument().TypeName)
		require.Equal(t, NormalConcrete, call.MustTypeArgument().Indirection)
		require.Equal(t, subpkg, call.MustTypeArgument().PkgPath)
		callArgs, err := call.ParseTypesFromArgs()
		require.NoError(t, err)
		require.NotEmpty(t, callArgs)
		require.Equal(t, subpkg, callArgs[0].PkgPath)
		require.Equal(t, "Type001", callArgs[0].TypeName)
		require.Equal(t, "Type002", callArgs[1].TypeName)
		require.Equal(t, "Type003", callArgs[2].TypeName)
		require.Equal(t, Pointer, callArgs[2].Indirection)
		require.Equal(t, "Type004", callArgs[3].TypeName)
		require.Equal(t, Pointer, callArgs[3].Indirection)
	})

	t.Run("Call number 10", func(t *testing.T) {
		call := ParseValueExprForMarkerFunctionCall(valueSpec(9))[0]
		require.Equal(t, MarkerFuncNewEnumType, call.CallExpr.MustIdentifyFunc().TypeName)
		require.Len(t, call.CallExpr.Args(), 0)
		require.NotNil(t, call.MustTypeArgument())
		require.Equal(t, "NiceEnumType", call.MustTypeArgument().TypeName)
		require.Equal(t, NormalConcrete, call.MustTypeArgument().Indirection)
		require.Equal(t, subpkg, call.MustTypeArgument().PkgPath)
	})
}

func printStuff(it any) {
	fmt.Printf("%T %#v\n", it, it)
}

var _ = printStuff

func nodePosition(pkg *decorator.Package, node dst.Node) token.Position {
	return pkg.Fset.Position(nodePos(pkg, node))
}
func nodePos(pkg *decorator.Package, node dst.Node) token.Pos {
	return pkg.Decorator.Map.Ast.Nodes[node].Pos()
}
