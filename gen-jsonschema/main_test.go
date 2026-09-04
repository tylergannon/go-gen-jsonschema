package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylergannon/go-gen-jsonschema/internal/builder"
)

func TestParseUnmarshalFormats(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"json", "both"} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			_, err := parseUnmarshalFormats(value)
			require.NoError(t, err)
		})
	}

	for _, value := range []string{"yaml", "toml"} {
		_, err := parseUnmarshalFormats(value)
		require.EqualError(t, err, `invalid --formats value "`+value+`": expected json or both`)
	}
}

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

func TestNewConfigValidationStubsFollowFormats(t *testing.T) {
	data, err := builder.RenderTemplate(configTmplContents, configArg{
		PkgName:  "example",
		BuildTag: "jsonschema",
		Validate: true,
		YAML:     true,
		Methods: []methodDef{
			{TypeName: "Example", MethodName: "Schema"},
		},
	})
	require.NoError(t, err)

	formatted, err := builder.FormatCodeWithGoimports(data.Bytes())
	require.NoError(t, err)
	source := string(formatted)
	require.Contains(t, source, "func (Example) ValidateJSON(")
	require.Contains(t, source, "func (Example) ValidateYAML(")
}
