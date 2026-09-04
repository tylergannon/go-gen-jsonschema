package syntax

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePackagePositionPreservesSpaces(t *testing.T) {
	t.Parallel()

	position, ok := parsePackagePosition("/tmp/a consumer/model.go:12:7")
	require.True(t, ok)
	require.Equal(t, "/tmp/a consumer/model.go", position.Filename)
	require.Equal(t, 12, position.Line)
	require.Equal(t, 7, position.Column)
}
