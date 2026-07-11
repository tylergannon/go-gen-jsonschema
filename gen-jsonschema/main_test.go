package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylergannon/go-gen-jsonschema/internal/builder"
)

func TestNewConfigUsesOnlyGoBuildConstraint(t *testing.T) {
	data, err := builder.RenderTemplate(configTmplContents, configArg{
		PkgName:  "example",
		BuildTag: "jsonschema",
		Methods: []methodDef{
			{TypeName: "Example", MethodName: "Schema"},
		},
	})
	require.NoError(t, err)

	formatted, err := builder.FormatCodeWithGoimports(data.Bytes())
	require.NoError(t, err)

	source := string(formatted)
	require.True(t, strings.HasPrefix(source, "//go:build jsonschema\n\npackage example\n"))
	require.NotContains(t, source, "// +build")
}
