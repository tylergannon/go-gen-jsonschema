package syntax

import (
	"bytes"
	"fmt"
	"go/token"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/decorator/resolver/gopackages"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/imports"
)

func TestFlattenTypes(t *testing.T) {
	pkgs, err := Load("./testfixtures/structtype")
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	require.Equal(t, "structtype", pkgs[0].Name)
	scanResult, err := loadPackageForTest(pkgs[0], "ExperimentRun", "ArrayOfSuperStruct")
	require.NoError(t, err)

	t.Run("should flatten types", func(t *testing.T) {
		require.Len(t, scanResult.LocalNamedTypes, 2)
	})

	t.Run("Should have a struct type for ExperimentRun", func(t *testing.T) {
		experimentRun := scanResult.LocalNamedTypes["ExperimentRun"]
		require.Equal(t, "ExperimentRun", experimentRun.Name())
		st, ok := experimentRun.Type().Expr().(*dst.StructType)
		require.True(t, ok)
		require.Len(t, st.Fields.List, 16)
	})

	t.Run("Should flatten the struct", func(t *testing.T) {
		experimentRun := scanResult.LocalNamedTypes["ExperimentRun"]
		st := NewStructType(experimentRun.Type().Expr().(*dst.StructType), experimentRun)
		flattened, err := st.Flatten(scanResult.Pkg.PkgPath, scanResult.resolveType, nil)
		require.NoError(t, err)
		file := &dst.File{Name: dst.NewIdent(scanResult.Pkg.Name)}
		ts := flattened.Concrete
		ts.Type = flattened.Expr
		file.Decls = append(file.Decls, &dst.GenDecl{Tok: token.TYPE, Specs: []dst.Spec{ts}})
		buf := bytes.Buffer{}
		printer := decorator.NewRestorerWithImports(
			"github.com/tylergannon/go-gen-jsonschema/internal/syntax/testfixtures/structtype",
			gopackages.New("./testfixtures/structtype"),
		)
		require.NoError(t, printer.Fprint(&buf, file))
		_, err = FormatCodeWithGoimports(buf.Bytes())
		require.NoError(t, err)
	})

	t.Run("Should stringify the struct", func(t *testing.T) {
		// intentionally left blank in original test
	})
}

func FormatCodeWithGoimports(source []byte) ([]byte, error) {
	options := &imports.Options{
		Comments:  true,
		TabIndent: true,
	}

	formatted, err := imports.Process("", source, options)
	if err != nil {
		return nil, fmt.Errorf("failed to format code: %w", err)
	}

	return formatted, nil
}
