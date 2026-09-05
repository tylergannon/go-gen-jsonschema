package polytype

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// advertisedRegenerableExamples are the examples/README.md entries that
// issue #77 fixed: each had no real go:generate entry point and no
// checked-in generated artifacts despite examples/README.md advertising
// them as regenerable. This test proves both examples now have a generation
// directive and the expected checked-in artifacts, so neither regression can
// silently return.
var advertisedRegenerableExamples = []struct {
	name      string
	artifacts []string
}{
	{
		name: "interfaces_options",
		artifacts: []string{
			"jsonschema_gen.go",
			"jsonschema/Owner.json",
			"jsonschema/Owner.json.sum",
		},
	},
	{
		name: "enums_stringmode",
		artifacts: []string{
			"jsonschema_gen.go",
			"jsonschema/Paint.json",
			"jsonschema/Paint.json.sum",
		},
	},
}

func TestAdvertisedExamplesHaveGenerateDirectiveAndArtifacts(t *testing.T) {
	repoRoot, err := os.Getwd()
	require.NoError(t, err)

	for _, example := range advertisedRegenerableExamples {
		name := example.name
		artifacts := example.artifacts
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join(repoRoot, "examples", name)
			requireGoGenerateDirective(t, dir, name)
			requireGeneratedArtifactsExist(t, dir, name, artifacts)
		})
	}
}

// requireGoGenerateDirective fails unless the example has a real
// //go:generate directive that invokes the generator.
func requireGoGenerateDirective(t *testing.T, dir, exampleName string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		require.NoError(t, err)
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "//go:generate") && strings.Contains(line, "polytype") {
				return
			}
		}
	}
	t.Fatalf("example %s has no //go:generate directive invoking polytype", exampleName)
}

// requireGeneratedArtifactsExist fails unless every expected artifact
// (jsonschema_gen.go, each schema JSON file, and each schema .sum file) is
// actually checked in, so removing any one of them - including just a
// .sum file - fails this test.
func requireGeneratedArtifactsExist(t *testing.T, dir, exampleName string, artifacts []string) {
	t.Helper()
	for _, rel := range artifacts {
		_, err := os.Stat(filepath.Join(dir, rel))
		require.NoError(t, err, "example %s is missing checked-in artifact %s", exampleName, rel)
	}
}
