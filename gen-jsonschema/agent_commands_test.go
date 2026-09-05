package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylergannon/go-gen-jsonschema/internal/inspection"
)

func TestVersionJSONWritesOnlyMachineResultToStdout(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	exitCode := runAgentCommand([]string{"version", "--json"}, &stdout, &stderr)

	require.Zero(t, exitCode)
	require.Empty(t, stderr.String())
	var result inspection.Result
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &result))
	require.Equal(t, "version", result.Kind)
	require.NotEqual(t, "latest", result.Tool.Version)
}

func TestInvalidInspectFlagStillWritesOneJSONResult(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	exitCode := runAgentCommand([]string{"inspect", "--json", "--not-a-flag"}, &stdout, &stderr)

	require.Equal(t, 2, exitCode)
	require.Empty(t, stderr.String())
	decoder := json.NewDecoder(&stdout)
	var result inspection.Result
	require.NoError(t, decoder.Decode(&result))
	require.False(t, decoder.More())
	require.Equal(t, inspection.StatusInvalid, result.Status)
	require.Equal(t, "invalid_request", result.Diagnostics[0].Code)
}

func TestInspectJSONReportsUnsupportedWithoutStdoutNoise(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join(".."))
	require.NoError(t, err)
	target := filepath.Join(repoRoot, "internal", "builder", "testfixtures", "structs")

	var stdout, stderr bytes.Buffer
	exitCode := runAgentCommand([]string{"inspect", "--json", "--target", target, "JSONTagNames"}, &stdout, &stderr)

	require.Equal(t, 3, exitCode)
	require.Empty(t, stderr.String())
	var result inspection.Result
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &result))
	require.Equal(t, inspection.StatusUnsupported, result.Status)
	require.Equal(t, "unsupported_required_omission", result.Types[0].Diagnostics[0].Code)
}

func TestInspectJSONClassifiesGenericModelWithoutStdoutNoise(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join(".."))
	require.NoError(t, err)
	target := filepath.Join(repoRoot, "internal", "builder", "testfixtures", "inspection_generic")

	var stdout, stderr bytes.Buffer
	exitCode := runAgentCommand([]string{"inspect", "--json", "--target", target, "Root"}, &stdout, &stderr)

	require.Equal(t, 3, exitCode)
	require.Empty(t, stderr.String())
	var result inspection.Result
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &result))
	require.Equal(t, inspection.StatusUnsupported, result.Status)
	require.Empty(t, result.Types)
	require.Equal(t, "unsupported_generic_type", result.Diagnostics[0].Code)
	require.Equal(t, inspection.ClassificationUnsupported, result.Diagnostics[0].Classification)
}

func TestInspectJSONReportsUnresolvedBuildContextAsUnknown(t *testing.T) {
	t.Setenv("GOFLAGS", `-tags='one two'`)

	var stdout, stderr bytes.Buffer
	exitCode := runAgentCommand([]string{"inspect", "--json", "Root"}, &stdout, &stderr)

	require.Equal(t, 3, exitCode)
	require.Empty(t, stderr.String())
	var result inspection.Result
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &result))
	require.Equal(t, inspection.StatusUnknown, result.Status)
	require.Equal(t, "build_context_unavailable", result.Diagnostics[0].Code)
	require.Equal(t, inspection.ClassificationUnknown, result.Diagnostics[0].Classification)
}

func TestInspectJSONUsesCustomTagsFromSavedGoEnv(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join(".."))
	require.NoError(t, err)
	target := filepath.Join(repoRoot, "internal", "builder", "testfixtures", "inspection_nested")
	goenv := filepath.Join(t.TempDir(), "go.env")
	require.NoError(t, os.WriteFile(goenv, []byte("GOFLAGS=-tags=custom\n"), 0o600))
	t.Setenv("GOENV", goenv)
	unsetEnv(t, "GOFLAGS")

	var stdout, stderr bytes.Buffer
	exitCode := runAgentCommand([]string{"inspect", "--json", "--target", target, "SavedTagHookModel"}, &stdout, &stderr)

	require.Equal(t, 3, exitCode)
	require.Empty(t, stderr.String())
	var result inspection.Result
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &result))
	require.Equal(t, inspection.StatusUnknown, result.Status)
	require.Equal(t, "unknown_custom_json_hook", result.Types[0].Diagnostics[0].Code)
	require.Equal(t, "custom_tag_hook.go", filepath.Base(result.Types[0].Diagnostics[0].Source.File))
}

func unsetEnv(t *testing.T, name string) {
	t.Helper()
	value, present := os.LookupEnv(name)
	require.NoError(t, os.Unsetenv(name))
	t.Cleanup(func() {
		if present {
			require.NoError(t, os.Setenv(name, value))
			return
		}
		require.NoError(t, os.Unsetenv(name))
	})
}

func TestHumanVersionUsesStderr(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	exitCode := runAgentCommand([]string{"version"}, &stdout, &stderr)

	require.Zero(t, exitCode)
	require.Empty(t, stdout.String())
	require.Contains(t, stderr.String(), "gen-jsonschema")
}

func TestAgentCommandHelpIsUsefulAndSuccessful(t *testing.T) {
	t.Parallel()

	var humanStdout, humanStderr bytes.Buffer
	exitCode := runAgentCommand([]string{"inspect", "--help"}, &humanStdout, &humanStderr)
	require.Zero(t, exitCode)
	require.Empty(t, humanStdout.String())
	require.Contains(t, humanStderr.String(), "Usage: gen-jsonschema inspect")
	require.Contains(t, humanStderr.String(), "--target DIR")

	var machineStdout, machineStderr bytes.Buffer
	exitCode = runAgentCommand([]string{"version", "--json", "--help"}, &machineStdout, &machineStderr)
	require.Zero(t, exitCode)
	require.Empty(t, machineStderr.String())
	var result inspection.Result
	require.NoError(t, json.Unmarshal(machineStdout.Bytes(), &result))
	require.Equal(t, "version", result.Kind)
	require.Contains(t, result.Usage, "Usage: gen-jsonschema version")
}

func TestAgentOperationPanicStillEmitsStructuredMachineError(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	exitCode := runAgentOperation("inspection", true, &stdout, &stderr, func() inspection.Result {
		panic("boom")
	})

	require.Equal(t, 1, exitCode)
	require.Empty(t, stderr.String())
	var result inspection.Result
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &result))
	require.Equal(t, "inspection", result.Kind)
	require.Equal(t, inspection.StatusError, result.Status)
	require.Equal(t, "internal_operation_panic", result.Diagnostics[0].Code)
}

func TestVersionInvalidRequestKeepsVersionKind(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	exitCode := runAgentCommand([]string{"version", "--json", "extra"}, &stdout, &stderr)

	require.Equal(t, 2, exitCode)
	require.Empty(t, stderr.String())
	var result inspection.Result
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &result))
	require.Equal(t, "version", result.Kind)
}
