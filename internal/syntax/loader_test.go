package syntax

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

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
