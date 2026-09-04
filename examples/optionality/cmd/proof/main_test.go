package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTranscriptMatchesExpected(t *testing.T) {
	result, err := run()
	require.NoError(t, err)

	actual, err := json.MarshalIndent(result, "", "  ")
	require.NoError(t, err)
	actual = append(actual, '\n')

	_, source, _, ok := runtime.Caller(0)
	require.True(t, ok)
	expected, err := os.ReadFile(filepath.Join(filepath.Dir(source), "..", "..", "proof", "expected.json"))
	require.NoError(t, err)
	require.Equal(t, string(expected), string(actual))
}
