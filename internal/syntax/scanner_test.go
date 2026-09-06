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
	pkgPath = "github.com/tylergannon/polytype/internal/syntax/testfixtures/typescanner"
	subpkg  = "github.com/tylergannon/polytype/internal/syntax/testfixtures/typescanner/scannersubpkg"
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
	require.Len(t, specs[0].specs, 6)
	valueSpec := func(idx int) ValueSpec {
		return NewValueSpec(specs[0].genDecl, specs[0].specs[idx].(*dst.ValueSpec), specs[0].pkg, specs[0].file)
	}

	var calls []MarkerFunctionCall
	for _, spec := range specs[0].specs {
		calls = append(calls, ParseValueExprForMarkerFunctionCall(NewValueSpec(specs[0].genDecl, spec.(*dst.ValueSpec), specs[0].pkg, specs[0].file))...)
	}
	require.Len(t, calls, 6)

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

	t.Run("Call number 7", func(t *testing.T) {
		call := ParseValueExprForMarkerFunctionCall(valueSpec(4))[0]
		require.Equal(t, MarkerFuncNewJSONSchemaBuilder, call.CallExpr.MustIdentifyFunc().TypeName)
		require.Len(t, call.CallExpr.Args(), 1)
		require.NotNil(t, call.MustTypeArgument())
		require.Equal(t, "TypeForSchemaFunction", call.MustTypeArgument().TypeName)
		require.Equal(t, NormalConcrete, call.MustTypeArgument().Indirection)
		require.Equal(t, subpkg, call.MustTypeArgument().PkgPath)
	})

	t.Run("Call number 8", func(t *testing.T) {
		call := ParseValueExprForMarkerFunctionCall(valueSpec(5))[0]
		require.Equal(t, MarkerFuncNewJSONSchemaBuilder, call.CallExpr.MustIdentifyFunc().TypeName)
		require.Len(t, call.CallExpr.Args(), 1)
		require.NotNil(t, call.MustTypeArgument())
		require.Equal(t, "PointerTypeForSchemaFunction", call.MustTypeArgument().TypeName)
		require.Equal(t, Pointer, call.MustTypeArgument().Indirection)
		require.Equal(t, subpkg, call.MustTypeArgument().PkgPath)
	})

}

func TestFluentDeclareParser(t *testing.T) {
	specs := LoadDecls("./testfixtures/typescanner", "fluent_calls.go", token.VAR)
	require.Len(t, specs, 1)
	require.Len(t, specs[0].specs, 5)
	valueSpec := func(idx int) ValueSpec {
		return NewValueSpec(specs[0].genDecl, specs[0].specs[idx].(*dst.ValueSpec), specs[0].pkg, specs[0].file)
	}
	localFuncs := loadPkgDecls(specs[0].pkg).funcDecls

	t.Run("full chain", func(t *testing.T) {
		calls := ParseValueExprForMarkerFunctionCall(valueSpec(0))
		require.Len(t, calls, 1)
		call := calls[0]
		require.Equal(t, MarkerFuncDeclare, call.CallExpr.MustIdentifyFunc().TypeName)

		method, isMethodRoot, err := call.ParseFluentDeclaration(localFuncs)
		require.NoError(t, err)
		require.True(t, isMethodRoot)
		require.Equal(t, "Schema", method.SchemaMethodName)
		require.Equal(t, pkgPath, method.Receiver.PkgPath)
		require.Equal(t, "FluentStruct", method.Receiver.TypeName)

		require.Equal(t, []SchemaMethodOptionInfo{
			{Kind: "WithStructAccessorMethod", FieldName: "A", ProviderName: "ASchema", ProviderIsMethod: true},
			{Kind: "WithStructFunctionMethod", FieldName: "B", ProviderName: "BSchema", ProviderIsMethod: true},
			{Kind: "WithFunction", FieldName: "C", ProviderName: "FluentBoolSchema"},
			{Kind: "WithStringerEnum", FieldName: "F"},
			{Kind: "AsRef"},
			{Kind: "WithRenderProviders"},
		}, method.Options)
	})

	t.Run("pointer-root chain", func(t *testing.T) {
		calls := ParseValueExprForMarkerFunctionCall(valueSpec(1))
		require.Len(t, calls, 1)
		method, isMethodRoot, err := calls[0].ParseFluentDeclaration(localFuncs)
		require.NoError(t, err)
		require.True(t, isMethodRoot)
		require.Equal(t, "PtrSchema", method.SchemaMethodName)
		require.Equal(t, pkgPath, method.Receiver.PkgPath)
		require.Equal(t, "FluentStruct", method.Receiver.TypeName)

		require.Equal(t, []SchemaMethodOptionInfo{
			{Kind: "WithStructAccessorMethod", FieldName: "A", ProviderName: "PtrASchema", ProviderIsMethod: true},
			{Kind: "WithStructFunctionMethod", FieldName: "B", ProviderName: "PtrBSchema", ProviderIsMethod: true},
		}, method.Options)
	})

	t.Run("free function root", func(t *testing.T) {
		calls := ParseValueExprForMarkerFunctionCall(valueSpec(2))
		require.Len(t, calls, 1)
		method, isMethodRoot, err := calls[0].ParseFluentDeclaration(localFuncs)
		require.NoError(t, err)
		require.False(t, isMethodRoot)
		require.Equal(t, "FluentFreeSchema", method.SchemaMethodName)
		require.Equal(t, pkgPath, method.Receiver.PkgPath)
		require.Equal(t, "FluentStruct", method.Receiver.TypeName)
		require.Empty(t, method.Options)
	})

	t.Run("bare Declare with no chain", func(t *testing.T) {
		calls := ParseValueExprForMarkerFunctionCall(valueSpec(3))
		require.Len(t, calls, 1)
		method, isMethodRoot, err := calls[0].ParseFluentDeclaration(localFuncs)
		require.NoError(t, err)
		require.True(t, isMethodRoot)
		require.Equal(t, "Schema", method.SchemaMethodName)
		require.Equal(t, "FluentStruct", method.Receiver.TypeName)
		require.Empty(t, method.Options)
	})

	t.Run("import alias root", func(t *testing.T) {
		calls := ParseValueExprForMarkerFunctionCall(valueSpec(4))
		require.Len(t, calls, 1)
		method, isMethodRoot, err := calls[0].ParseFluentDeclaration(localFuncs)
		require.NoError(t, err)
		require.True(t, isMethodRoot)
		require.Equal(t, "Schema", method.SchemaMethodName)
		require.Equal(t, subpkg, method.Receiver.PkgPath)
		require.Equal(t, "TypeForSchemaMethod", method.Receiver.TypeName)
		require.Equal(t, []SchemaMethodOptionInfo{{Kind: "AsRef"}}, method.Options)
	})
}

// TestFluentChainFieldSelectorMismatchFailsToLoad proves that a fluent chain
// link whose field selector names a type other than the Declare(...) root
// (a typo Go's type system can't catch, since StringerEnum's field parameter is
// `any`) is a hard, source-positioned scanner error rather than the silent
// skip the legacy variadic-option parser applies to an analogous mismatch.
func TestFluentChainFieldSelectorMismatchFailsToLoad(t *testing.T) {
	pkgs, err := Load("./testfixtures/fluent_field_mismatch")
	require.NoError(t, err)
	require.Len(t, pkgs, 1)

	_, err = LoadPackage(pkgs[0])
	require.Error(t, err)
	require.ErrorContains(t, err, "polytype.Declare: .StringerEnum expects a field selector on Owner{}")
	require.ErrorContains(t, err, "fixture.go")
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
