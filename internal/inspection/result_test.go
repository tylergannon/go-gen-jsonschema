package inspection

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAggregateStatusPrecedence(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name        string
		diagnostics []Diagnostic
		want        Status
	}{
		{name: "supported", want: StatusSupported},
		{name: "unknown", diagnostics: []Diagnostic{{Classification: ClassificationUnknown}}, want: StatusUnknown},
		{name: "unsupported_over_unknown", diagnostics: []Diagnostic{{Classification: ClassificationUnknown}, {Classification: ClassificationUnsupported}}, want: StatusUnsupported},
		{name: "invalid_over_unsupported", diagnostics: []Diagnostic{{Classification: ClassificationUnsupported}, {Classification: ClassificationInvalidRequest}}, want: StatusInvalid},
		{name: "error_over_all", diagnostics: []Diagnostic{{Classification: ClassificationInvalidRequest}, {Classification: ClassificationInternal}}, want: StatusError},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, test.want, AggregateStatus(test.diagnostics))
		})
	}
}

func TestVersionUsesRuntimeBuildIdentity(t *testing.T) {
	t.Parallel()

	result := Version()
	require.Equal(t, "version", result.Kind)
	require.Equal(t, SchemaVersion, result.SchemaVersion)
	require.Equal(t, ContractVersion, result.ContractVersion)
	require.NotEmpty(t, result.Tool.Version)
	require.NotEqual(t, "latest", result.Tool.Version)
	require.Contains(t, []string{"development", "pseudo", "release", "unknown"}, result.Tool.VersionState)
	require.NotEmpty(t, result.Tool.Revision)
	require.Contains(t, []string{"known", "unknown"}, result.Tool.RevisionState)
	require.NotEmpty(t, result.Capabilities)
}
