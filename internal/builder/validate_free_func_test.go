package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// writeMultiFileFixture is writeTypeGrammarFixture's multi-file sibling, for
// cases that need types and jsonschema-tagged registrations split across
// files the way real registration code is (types.go untagged, schema.go
// //go:build jsonschema).
func writeMultiFileFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	dir := t.TempDir()
	module := fmt.Sprintf("module example.com/typegrammarfixture\n\ngo 1.27\n\nrequire github.com/tylergannon/polytype v0.0.0\nreplace github.com/tylergannon/polytype => %s\n", root)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(module), 0o644))
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
	}
	return dir
}

// TestValidateRejectsFreeFunctionPointerRoot proves that --validate fails
// fast, with an actionable error, instead of silently succeeding while
// omitting ValidateJSON for a free-function schema root whose type can't
// have a method declared on it (pointer/interface underlying type). Before
// this check, generation would succeed and simply produce no ValidateJSON
// for that type -- a silent partial result.
func TestValidateRejectsFreeFunctionPointerRoot(t *testing.T) {
	dir := writeTypeGrammarFixture(t, `//go:build jsonschema

package fixture

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

type PointerRoot *int

func PointerRootSchema(PointerRoot) json.RawMessage { panic("not implemented") }

var _ = polytype.Declare(PointerRootSchema)
`)

	err := Run(BuilderArgs{TargetDir: dir, Validate: true})
	require.ErrorContains(t, err, "--validate cannot generate ValidateJSON for PointerRoot")
}

// TestFreeFunctionRootForRegisteredInterfaceCompilesAndRuns proves that a
// free-function schema root for a type registered via the legacy
// NewInterfaceImpl (recorded in Scan.Interfaces, not Scan.LocalNamedTypes)
// is correctly classified as needing a free function, not a method: before
// this fix, hasInvalidMethodReceiverBase only consulted LocalNamedTypes, so
// this exact shape would be misrouted into SchemaMethods() and generate an
// uncompilable `func (Value) ValueSchema()` (Go forbids an interface
// receiver base). Runs generation for real and calls the result.
func TestFreeFunctionRootForRegisteredInterfaceCompilesAndRuns(t *testing.T) {
	dir := writeMultiFileFixture(t, map[string]string{
		"types.go": `package fixture

type Value interface{ value() }

type First struct {
	Name string ` + "`json:\"name\"`" + `
}

func (First) value() {}
`,
		"schema.go": `//go:build jsonschema

package fixture

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func ValueSchema(Value) json.RawMessage { panic("not implemented") }

var (
	_ = polytype.NewInterfaceImpl[Value](First{})
	_ = polytype.Declare(ValueSchema)
)
`,
	})

	require.NoError(t, Run(BuilderArgs{TargetDir: dir}))

	generated, err := os.ReadFile(filepath.Join(dir, "jsonschema_gen.go"))
	require.NoError(t, err)
	require.Contains(t, string(generated), "func ValueSchema(Value) json.RawMessage {")
	require.NotContains(t, string(generated), "func (Value) ValueSchema()")
}

// TestRenderProvidersRejectsFreeFunctionPointerRoot proves that
// RenderProviders() fails fast, with an actionable error, for a
// free-function root whose type can't have a method (so no RenderedSchema()
// could ever be generated for it) instead of silently writing a
// `.json.tmpl` that nothing can execute.
func TestRenderProvidersRejectsFreeFunctionPointerRoot(t *testing.T) {
	dir := writeTypeGrammarFixture(t, `//go:build jsonschema

package fixture

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

type PointerRoot *int

func PointerRootSchema(PointerRoot) json.RawMessage { panic("not implemented") }

var _ = polytype.Declare(PointerRootSchema).RenderProviders()
`)

	err := Run(BuilderArgs{TargetDir: dir})
	require.ErrorContains(t, err, "PointerRoot: RenderProviders() is not supported")
}
