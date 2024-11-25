package testutils

import (
	"fmt"
	"os/exec"

	"github.com/onsi/gomega/types"
)

// MatchGoldenFile compares two files using `diff` for reporting differences.
func MatchGoldenFile() types.GomegaMatcher {
	return &goldenFileMatcher{}
}

type goldenFileMatcher struct {
	actualFile string
	diffOutput string
}

func (g *goldenFileMatcher) Match(actual interface{}) (success bool, err error) {
	actualFile, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("MatchGoldenFile expects a file path (string) as actual, got %T", actual)
	}
	g.actualFile = actualFile

	expectedFile := actualFile + "lden"

	// Run diff to check for differences
	cmd := exec.Command("diff", "-u", g.actualFile, expectedFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Non-zero exit code from diff indicates differences
		g.diffOutput = string(output)
		return false, nil
	}

	// No differences
	g.diffOutput = ""
	return true, nil
}

func (g *goldenFileMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Golden file comparison failed:\n%s", g.diffOutput)
}

func (g *goldenFileMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return "Golden file unexpectedly matched"
}
