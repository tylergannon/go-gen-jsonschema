package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylergannon/go-gen-jsonschema/internal/builder"
)

func TestParseUnmarshalFormats(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"json", "yaml", "both"} {
		value := value
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			_, err := parseUnmarshalFormats(value)
			require.NoError(t, err)
		})
	}

	_, err := parseUnmarshalFormats("toml")
	require.EqualError(t, err, `invalid --formats value "toml": expected json, yaml, or both`)
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
