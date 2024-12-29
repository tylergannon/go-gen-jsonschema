package testutils

import (
	"fmt"
	"github.com/onsi/gomega/matchers"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/onsi/gomega/types"
)

const defaultExt = "lden"

// MatchGoldenFile compares two files using `diff` for reporting differences.
func MatchGoldenFile(ext ...string) types.GomegaMatcher {
	if len(ext) == 0 {
		ext = append(ext, defaultExt)
	}
	return &goldenFileMatcher{
		extension: ext[0],
	}
}

type goldenFileMatcher struct {
	extension          string
	actualFile         string
	diffOutput         string
	matcher            *matchers.MatchJSONMatcher
	actualFileContents []byte
}

func (g *goldenFileMatcher) Match(actual interface{}) (success bool, err error) {
	actualFile, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("MatchGoldenFile expects a file path (string) as actual, got %T", actual)
	}
	g.actualFile = actualFile

	expectedFile := actualFile + g.extension

	if strings.ToLower(filepath.Ext(actualFile)) == ".json" {
		actualFileContents, err := os.ReadFile(actualFile)
		if err != nil {
			return false, fmt.Errorf("failed to read file %s: %w", actualFile, err)
		}
		expectedFileContents, err := os.ReadFile(expectedFile)
		if err != nil {
			return false, fmt.Errorf("failed to read file %s: %w", expectedFile, err)
		}

		g.matcher = &matchers.MatchJSONMatcher{
			JSONToMatch: expectedFileContents,
		}
		g.actualFileContents = actualFileContents
		match, err := g.matcher.Match(actualFileContents)
		if match {
			return true, nil
		}
		return false, err
	}

	// Run diff to check for differences
	cmd := exec.Command("diff", "-u", expectedFile, g.actualFile)
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
	if g.matcher == nil {
		return fmt.Sprintf("Golden file comparison failed:\n%s", g.diffOutput)
	}
	return g.matcher.FailureMessage(g.actualFileContents)
}

func (g *goldenFileMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	if g.matcher == nil {
		return "Golden file unexpectedly matched"
	}
	return g.matcher.NegatedFailureMessage(g.actualFileContents)
}
