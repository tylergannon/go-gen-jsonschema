package main

import (
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylergannon/go-gen-jsonschema/internal/builder"
	"github.com/tylergannon/go-gen-jsonschema/internal/testutils"
)

func TestGenTypeScriptFlags(t *testing.T) {
	t.Parallel()

	flags, options := newGenFlagSet(flag.ContinueOnError)
	err := flags.Parse([]string{
		"--typescript", "web/generated",
		"--typescript-barrel",
	})
	require.NoError(t, err)
	require.Equal(t, "web/generated", options.typeScriptDir)
	require.True(t, options.typeScriptBarrel)
}

func TestGenTypeScriptFlagsDefaultToDisabled(t *testing.T) {
	t.Parallel()

	flags, options := newGenFlagSet(flag.ContinueOnError)
	require.NoError(t, flags.Parse(nil))
	require.Empty(t, options.typeScriptDir)
	require.False(t, options.typeScriptBarrel)
}

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

func TestNewConfigUsesFluentDeclareForm(t *testing.T) {
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
	require.Contains(t, source, "jsonschema.Declare(Example.Schema)")
	require.NotContains(t, source, "NewJSONSchemaMethod")
}

// TestGenCommandRejectsInvalidFluentFieldAssociationWithSourcePosition runs
// the actual gen-jsonschema binary (not an in-process builder.Run call)
// against the checked-in fluent_field_mismatch scanner fixture
// (internal/syntax/testfixtures/fluent_field_mismatch), whose .Enum(...)
// chain link names a field on a type other than its Declare(...) root. This
// proves the command itself - not just the internal scanner/builder APIs -
// fails with a non-zero exit and a source-positioned diagnostic naming the
// offending file for an invalid fluent chain.
func TestGenCommandRejectsInvalidFluentFieldAssociationWithSourcePosition(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	exitCode, _, stderr, err := testutils.RunCommand("go", cwd, "run", ".", "-target", "../internal/syntax/testfixtures/fluent_field_mismatch")
	require.NoError(t, err)
	require.NotEqual(t, 0, exitCode, "stderr:\n%s", stderr)
	require.Contains(t, stderr, "jsonschema.Declare: .Enum expects a field selector on Owner{}")
	require.Contains(t, stderr, "fluent_field_mismatch/fixture.go")
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
