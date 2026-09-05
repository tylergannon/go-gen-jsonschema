package syntax

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProductionJSONMethodsPreserveCustomTagsAndExcludeGeneratorTag(t *testing.T) {
	t.Setenv("GOFLAGS", "-p=2 -tags=custom,"+BuildTag)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "model.go"), []byte(`package model

type Model struct{}
type Stub struct{}
`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "custom_hook.go"), []byte(`//go:build custom

package model

func (Model) MarshalJSON() ([]byte, error) { return nil, nil }
`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "schema.go"), []byte(`//go:build jsonschema

package model

func (Stub) MarshalJSON() ([]byte, error) { return nil, nil }
`), 0o600))

	context, err := ResolveBuildContext()
	require.NoError(t, err)
	methods, err := context.FindProductionJSONMethods(dir, nil)
	require.NoError(t, err)
	require.Len(t, methods, 1)
	require.Equal(t, "Model", methods[0].Receiver)
	require.Equal(t, "MarshalJSON", methods[0].Name)
}

func TestGenerationBuildTagsAddReservedTagAndPreserveCustomTags(t *testing.T) {
	goenv := filepath.Join(t.TempDir(), "go.env")
	require.NoError(t, os.WriteFile(goenv, []byte("GOFLAGS='-tags=second custom jsonschema'\nGOOS=linux\nGOARCH=amd64\nCGO_ENABLED=0\n"), 0o600))
	environment := withoutEnvironment(os.Environ(), "GOFLAGS", "GOENV", "GOOS", "GOARCH", "CGO_ENABLED")
	environment = append(environment, "GOENV="+goenv)

	context, err := resolveBuildContext(environment)
	require.NoError(t, err)

	require.Equal(t, []string{"custom", "jsonschema", "second"}, context.generationTags)
	require.Equal(t, []string{"custom", "second"}, context.production.BuildTags)
	require.Equal(t, "linux", context.production.GOOS)
	require.Equal(t, "amd64", context.production.GOARCH)
	require.False(t, context.production.CgoEnabled)
}

func TestMalformedEffectiveGOFLAGSDoesNotSilentlyDropTags(t *testing.T) {
	_, err := parseGOFLAGSTags(`-p=2 -tags='one two'`)

	require.ErrorContains(t, err, "non-flag")
}

func withoutEnvironment(environment []string, names ...string) []string {
	filtered := make([]string, 0, len(environment))
	for _, entry := range environment {
		name, _, _ := strings.Cut(entry, "=")
		if !slices.Contains(names, name) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}
