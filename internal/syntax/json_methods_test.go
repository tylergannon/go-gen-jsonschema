package syntax

import (
	"os"
	"path/filepath"
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

	methods, err := FindProductionJSONMethods(dir, nil)
	require.NoError(t, err)
	require.Len(t, methods, 1)
	require.Equal(t, "Model", methods[0].Receiver)
	require.Equal(t, "MarshalJSON", methods[0].Name)
}

func TestGenerationBuildTagsAddReservedTagAndPreserveCustomTags(t *testing.T) {
	t.Setenv("GOFLAGS", `-p=2 -tags "second,custom,jsonschema"`)

	require.Equal(t, []string{"custom", "jsonschema", "second"}, generationBuildTags())
	require.Equal(t, []string{"custom", "second"}, productionBuildTags())
}
