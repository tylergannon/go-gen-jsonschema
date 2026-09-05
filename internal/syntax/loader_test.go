package syntax

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadUsesTargetAsPackageWorkingDirectory(t *testing.T) {
	t.Parallel()

	originalConfigDir := DefaultPackageCfg.Dir
	target := filepath.Join(t.TempDir(), "independent-module")
	require.NoError(t, os.MkdirAll(target, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(target, "go.mod"),
		[]byte("module example.com/independent\n\ngo 1.24\n"),
		0o644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(target, "types.go"),
		[]byte("package independent\n\ntype Example struct { Name string `json:\"name\"` }\n"),
		0o644,
	))

	packages, err := Load(target)
	require.NoError(t, err)
	require.Len(t, packages, 1)
	require.Equal(t, "independent", packages[0].Name)
	require.Equal(t, "example.com/independent", packages[0].PkgPath)
	require.Equal(t, originalConfigDir, DefaultPackageCfg.Dir, "Load must not mutate the shared default config")
}
