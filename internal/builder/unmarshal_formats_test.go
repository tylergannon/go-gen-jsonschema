package builder_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylergannon/go-gen-jsonschema/internal/testutils"
)

func TestGeneratedUnmarshalFormats(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	repoRoot := filepath.Clean(filepath.Join(cwd, "..", ".."))
	generator := filepath.Join(t.TempDir(), "gen-jsonschema")
	exitCode, stdout, stderr, err := testutils.RunCommand("go", repoRoot, "build", "-o", generator, "./gen-jsonschema")
	require.NoError(t, err)
	require.Equal(t, 0, exitCode, "stdout:\n%s\nstderr:\n%s", stdout, stderr)

	tests := []struct {
		name     string
		formats  string
		wantJSON bool
		wantYAML bool
	}{
		{name: "default", wantJSON: true},
		{name: "json", formats: "json", wantJSON: true},
		{name: "both", formats: "both", wantJSON: true, wantYAML: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetDir, err := os.MkdirTemp("test_run", "formats-")
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, os.RemoveAll(targetDir)) })
			targetDir, err = filepath.Abs(targetDir)
			require.NoError(t, err)
			fixtureDir := filepath.Join("testfixtures", "interfaces")
			require.NoError(t, testutils.CopyDir(fixtureDir, targetDir))

			var args []string
			if tt.formats != "" {
				args = append(args, "--formats="+tt.formats)
			}
			exitCode, stdout, stderr, err := testutils.RunCommand(generator, targetDir, args...)
			require.NoError(t, err)
			require.Equal(t, 0, exitCode, "stdout:\n%s\nstderr:\n%s", stdout, stderr)

			generated, err := os.ReadFile(filepath.Join(targetDir, "jsonschema_gen.go"))
			require.NoError(t, err)
			source := string(generated)

			if tt.wantJSON {
				require.Contains(t, source, "func (f *FancyStruct) UnmarshalJSON(")
				require.Contains(t, source, "func __jsonUnmarshal__interfaces__TestInterface__")
			} else {
				require.NotContains(t, source, "UnmarshalJSON(")
				require.NotContains(t, source, "__jsonUnmarshal__")
			}

			if tt.wantYAML {
				require.Contains(t, source, "go.yaml.in/yaml/v4")
				require.Contains(t, source, "func (f *FancyStruct) UnmarshalYAML(")
				require.Contains(t, source, "func __gen_jsonschema_yamlNodeToJSON(")
				require.NotContains(t, source, "__yamlUnmarshal__")
			} else {
				require.NotContains(t, source, "go.yaml.in/yaml/v4")
				require.NotContains(t, source, "UnmarshalYAML(")
				require.NotContains(t, source, "__gen_jsonschema_yamlNodeToJSON(")
			}

			exitCode, stdout, stderr, err = testutils.RunCommand("go", targetDir, "build", "./...")
			require.NoError(t, err)
			require.Equal(t, 0, exitCode, "stdout:\n%s\nstderr:\n%s", stdout, stderr)
		})
	}
}
