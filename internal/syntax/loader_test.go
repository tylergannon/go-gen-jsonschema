package syntax

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestRemoteLoadUsesOriginalTargetModuleFromOutsideWorkingDirectory(t *testing.T) {
	target := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(target, "left"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(target, "go.mod"), []byte("module example.com/target\n\ngo 1.24\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(target, "types.go"), []byte(`package target

import "example.com/target/left"

type Root struct { Value left.Value }
`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(target, "left", "types.go"), []byte(`package left

type Value struct { Name string }
`), 0o600))

	context, err := ResolveBuildContext()
	require.NoError(t, err)
	packages, err := LoadTargetWithBuildContext(target, context, true)
	require.NoError(t, err)
	require.Len(t, packages, 1)

	scan := newScanResultWithBuildContext(packages[0], map[string]ScanResult{}, &context)
	scan.loadDir = target
	scan.loadReadonly = true
	remote, err := scan.loadRemotePackage("example.com/target/left")
	require.NoError(t, err)
	require.Len(t, remote, 1)
	require.Equal(t, "example.com/target/left", remote[0].PkgPath)
}

func TestTargetLoaderValidatesTheSameSourceInReadonlyAndWritableModes(t *testing.T) {
	target := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(target, "go.mod"), []byte("module example.com/invalid\n\ngo 1.24\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(target, "one.go"), []byte("package invalid\n\nfunc Duplicate() {}\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(target, "two.go"), []byte("package invalid\n\nfunc Duplicate() {}\n"), 0o600))
	context, err := ResolveBuildContext()
	require.NoError(t, err)

	for _, readonly := range []bool{false, true} {
		_, err := LoadTargetWithBuildContext(target, context, readonly)
		var loadErr *PackageLoadError
		require.ErrorAs(t, err, &loadErr)
		require.True(t, loadErr.HasSourceError())
	}
}

func TestTargetLoaderReturnsTypedNoGoPackageError(t *testing.T) {
	target := t.TempDir()
	context, err := ResolveBuildContext()
	require.NoError(t, err)

	_, err = LoadTargetWithBuildContext(target, context, true)
	var noGoPackage *NoGoPackageError
	require.ErrorAs(t, err, &noGoPackage)
	require.Equal(t, target, noGoPackage.TargetDir)
}

func TestParsePackagePositionPreservesSpaces(t *testing.T) {
	t.Parallel()

	position, ok := parsePackagePosition("/tmp/a consumer/model.go:12:7")
	require.True(t, ok)
	require.Equal(t, "/tmp/a consumer/model.go", position.Filename)
	require.Equal(t, 12, position.Line)
	require.Equal(t, 7, position.Column)
}

func TestPackageLoadErrorDistinguishesToolchainFromSource(t *testing.T) {
	t.Parallel()

	loadErr := &PackageLoadError{MissingImport: true, Errors: []packages.Error{
		{Kind: packages.TypeError, Pos: "/tmp/model.go:3:8", Msg: "could not import dependency"},
		{Kind: packages.ListError, Msg: "missing go.sum entry"},
	}}
	require.True(t, loadErr.HasSourceError())
	require.True(t, loadErr.HasToolchainError())
}
