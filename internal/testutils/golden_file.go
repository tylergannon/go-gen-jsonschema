package testutils

import (
	"os/exec"
	"testing"
)

const defaultExt = ".golden"

// AssertGoldenFile compares a file against its golden version using diff.
func AssertGoldenFile(t *testing.T, actualFile string, ext ...string) {
	t.Helper()
	extension := defaultExt
	if len(ext) > 0 {
		extension = ext[0]
	}
	expectedFile := actualFile + extension
	cmd := exec.Command("diff", "-u", expectedFile, actualFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Golden file comparison failed:\n%s", string(output))
	}
}
