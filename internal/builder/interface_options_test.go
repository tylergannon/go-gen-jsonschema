package builder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylergannon/polytype/internal/syntax"
)

func TestInlineInterfaceRegistrationDiagnostics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		options   string
		wantError string
	}{
		{
			name:      "missing implementations",
			options:   `polytype.WithInterface(Owner{}.Value, polytype.Discriminator("kind"))`,
			wantError: "missing interface implementations",
		},
		{
			name:      "duplicate wire value",
			options:   `polytype.WithInterface(Owner{}.Value, polytype.Impl("same", First{}), polytype.Impl("same", Second{}))`,
			wantError: `duplicate discriminator value "same"`,
		},
		{
			name:      "duplicate implementation",
			options:   `polytype.WithInterface(Owner{}.Value, polytype.Impl("first", First{}), polytype.Impl("again", First{}))`,
			wantError: "duplicate discriminator registration",
		},
		{
			name: "mixed registration forms",
			options: `polytype.WithInterface(Owner{}.Value, polytype.Impl("first", First{})),
	polytype.WithInterfaceImpls(Owner{}.Value, First{})`,
			wantError: "cannot combine Impl(...) options with WithInterfaceImpls",
		},
		{
			name:      "implementation does not satisfy interface",
			options:   `polytype.WithInterface(Owner{}.Value, polytype.Impl("stranger", Stranger{}))`,
			wantError: "does not implement Value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			targetDir := writeInlineInterfaceFixture(t, tc.options)
			pkgs, err := syntax.Load(targetDir)
			require.NoError(t, err)
			require.Len(t, pkgs, 1)
			require.Empty(t, pkgs[0].Errors)

			_, err = New(pkgs[0])
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func writeInlineInterfaceFixture(t *testing.T, options string) string {
	t.Helper()

	cwd, err := os.Getwd()
	require.NoError(t, err)
	targetDir, err := os.MkdirTemp(filepath.Join(cwd, "testfixtures"), "inline_interface_")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(targetDir))
	})

	source := `//go:build jsonschema

package fixture

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

type Value interface{ value() }

type First struct { Name string ` + "`json:\"name\"`" + ` }
func (First) value() {}

type Second struct { Count int ` + "`json:\"count\"`" + ` }
func (Second) value() {}

type Stranger struct { Enabled bool ` + "`json:\"enabled\"`" + ` }

type Owner struct { Value Value ` + "`json:\"value\"`" + ` }
func (Owner) Schema() json.RawMessage { panic("not implemented") }

var _ = polytype.NewJSONSchemaMethod(
	Owner.Schema,
	` + options + `,
)
`
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "schema.go"), []byte(source), 0o644))
	return targetDir
}
