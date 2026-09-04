package inspection

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInspectSupportedAndUnregisteredTypesIndependently(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	target := filepath.Join(repoRoot, "internal", "builder", "testfixtures", "basictypes")

	result := Inspect(InspectRequest{
		TargetDir: target,
		TypeNames: []string{"Missing", "TypeInItsOwnDecl"},
	})

	require.Equal(t, StatusUnknown, result.Status)
	require.Len(t, result.Types, 2)
	require.Equal(t, "github.com/tylergannon/go-gen-jsonschema/internal/builder/testfixtures/basictypes.Missing", result.Types[0].TypePath)
	require.Equal(t, StatusUnknown, result.Types[0].Status)
	require.Equal(t, "type_not_registered", result.Types[0].Diagnostics[0].Code)
	require.Equal(t, "github.com/tylergannon/go-gen-jsonschema/internal/builder/testfixtures/basictypes.TypeInItsOwnDecl", result.Types[1].TypePath)
	require.Equal(t, StatusSupported, result.Types[1].Status)
	for _, capability := range result.Types[1].Capabilities {
		require.Equal(t, StatusSupported, capability.Status, capability.Name)
	}
}

func TestInspectReportsOrdinaryOmissionWithSourceContext(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	target := filepath.Join(repoRoot, "internal", "builder", "testfixtures", "structs")

	result := Inspect(InspectRequest{TargetDir: target, TypeNames: []string{"JSONTagNames"}})

	require.Equal(t, StatusUnsupported, result.Status)
	require.Len(t, result.Types, 1)
	require.Equal(t, StatusUnsupported, result.Types[0].Status)
	require.NotEmpty(t, result.Types[0].Diagnostics)
	diagnostic := result.Types[0].Diagnostics[0]
	require.Equal(t, "unsupported_required_omission", diagnostic.Code)
	require.Contains(t, diagnostic.FieldPath, "JSONTagNames.")
	require.NotNil(t, diagnostic.Source)
	require.Equal(t, "struct_types.go", filepath.Base(diagnostic.Source.File))
}

func TestInspectProviderReportsPerSurfaceCapabilities(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	target := filepath.Join(repoRoot, "examples", "providers_rendering")

	result := Inspect(InspectRequest{TargetDir: target, TypeNames: []string{"Example"}})

	require.Equal(t, StatusUnsupported, result.Status)
	require.Len(t, result.Types, 1)
	capabilities := make(map[string]Status)
	for _, capability := range result.Types[0].Capabilities {
		capabilities[capability.Name] = capability.Status
	}
	require.Equal(t, StatusSupported, capabilities["schema"])
	require.Equal(t, StatusUnsupported, capabilities["json_encode"])
	require.Equal(t, StatusUnsupported, capabilities["json_decode"])
	require.Equal(t, StatusUnsupported, capabilities["validation"])
}

func TestInspectPropagatesNestedCodecRequirements(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	target := filepath.Join(repoRoot, "internal", "builder", "testfixtures", "inspection_nested")

	result := Inspect(InspectRequest{TargetDir: target, TypeNames: []string{"Parent"}})

	require.Equal(t, StatusUnsupported, result.Status)
	require.Len(t, result.Types, 1)
	capabilities := make(map[string]Status)
	for _, capability := range result.Types[0].Capabilities {
		capabilities[capability.Name] = capability.Status
	}
	require.Equal(t, StatusSupported, capabilities["schema"])
	require.Equal(t, StatusUnsupported, capabilities["json_encode"])
	require.Equal(t, StatusUnsupported, capabilities["json_decode"])
	for _, diagnostic := range result.Types[0].Diagnostics {
		require.NotEqual(t, "unsupported_inline_interface", diagnostic.Code)
	}
}

func TestInspectReportsUnknownCustomHookAndWireMismatches(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	target := filepath.Join(repoRoot, "internal", "builder", "testfixtures", "inspection_nested")

	hookResult := Inspect(InspectRequest{TargetDir: target, TypeNames: []string{"HookModel"}})
	require.Equal(t, StatusUnknown, hookResult.Status)
	require.Equal(t, "unknown_custom_json_hook", hookResult.Types[0].Diagnostics[0].Code)
	require.Equal(t, "HookModel.Hook", hookResult.Types[0].Diagnostics[0].FieldPath)
	require.Equal(t, "production_hook.go", filepath.Base(hookResult.Types[0].Diagnostics[0].Source.File))

	mismatchResult := Inspect(InspectRequest{TargetDir: target, TypeNames: []string{"WireMismatch"}})
	require.Equal(t, StatusUnsupported, mismatchResult.Status)
	codes := make(map[string]bool)
	for _, diagnostic := range mismatchResult.Types[0].Diagnostics {
		codes[diagnostic.Code] = true
	}
	require.True(t, codes["unsupported_base64_bytes"])
	require.True(t, codes["unsupported_json_string"])
	require.Contains(t, diagnosticFieldPaths(mismatchResult), "WireMismatch.Bytes")
	require.Contains(t, diagnosticFieldPaths(mismatchResult), "WireMismatch.Count")
	require.Contains(t, diagnosticFieldPaths(mismatchResult), "WireMismatch.Aliased")
	require.Contains(t, diagnosticFieldPaths(mismatchResult), "WireMismatch.Inline.Count")
	require.NotContains(t, diagnosticFieldPaths(mismatchResult), "WireMismatch.Inline.Ignored")
}

func TestInspectProviderWithoutRenderOptionStillReportsLimits(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	target := filepath.Join(repoRoot, "internal", "builder", "testfixtures", "inspection_nested")

	result := Inspect(InspectRequest{TargetDir: target, TypeNames: []string{"ProviderModel"}})

	require.Equal(t, StatusUnsupported, result.Status)
	require.Equal(t, "provider_codec_unavailable", result.Types[0].Diagnostics[0].Code)
}

func TestInspectDoesNotAdvertiseArbitraryExternalType(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	target := filepath.Join(repoRoot, "internal", "builder", "testfixtures", "inspection_nested")

	result := Inspect(InspectRequest{TargetDir: target, TypeNames: []string{"ExternalModel"}})

	require.Equal(t, StatusUnknown, result.Status)
	require.Equal(t, "unknown_external_type", result.Types[0].Diagnostics[0].Code)
	require.Equal(t, "ExternalModel.URL", result.Types[0].Diagnostics[0].FieldPath)
}

func TestInspectUnsupportedUnionContainerHasStructuredFieldContext(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	target := filepath.Join(repoRoot, "internal", "builder", "testfixtures", "inspection_nested")

	result := Inspect(InspectRequest{TargetDir: target, TypeNames: []string{"BadUnion"}})

	require.Equal(t, StatusUnsupported, result.Status)
	var found *Diagnostic
	for index := range result.Types[0].Diagnostics {
		if result.Types[0].Diagnostics[index].Code == "unsupported_interface_shape" {
			found = &result.Types[0].Diagnostics[index]
		}
	}
	require.NotNil(t, found)
	require.Equal(t, "BadUnion.Events", found.FieldPath)
	require.Equal(t, "types.go", filepath.Base(found.Source.File))
}

func TestInspectExcludesGenerationOnlyHooksAndQualifiesProductionHookTypes(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	target := filepath.Join(repoRoot, "internal", "builder", "testfixtures", "inspection_nested")

	result := Inspect(InspectRequest{TargetDir: target, TypeNames: []string{"RemoteSameName", "StubOnly"}})

	require.Equal(t, StatusSupported, result.Status)
	require.Len(t, result.Types, 2)
	for _, inspectedType := range result.Types {
		require.Equal(t, StatusSupported, inspectedType.Status, inspectedType.TypePath)
		require.Empty(t, inspectedType.Diagnostics)
	}
}

func diagnosticFieldPaths(result Result) []string {
	var paths []string
	for _, diagnostic := range result.Types[0].Diagnostics {
		paths = append(paths, diagnostic.FieldPath)
	}
	return paths
}
