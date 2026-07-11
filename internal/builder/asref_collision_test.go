package builder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylergannon/go-gen-jsonschema/internal/syntax"
)

// TestAsRefDefinitionNameCollisionFailsDuringGeneration proves that two
// distinct AsRef()'d types which resolve to the same bare "$defs" name
// (here: two different "Shared" types, one local and one imported under a
// selector-expression receiver) are rejected as a hard, generation-time
// error rather than silently colliding in the generated "$defs" map.
func TestAsRefDefinitionNameCollisionFailsDuringGeneration(t *testing.T) {
	t.Parallel()

	depDir := writeAsRefCollisionDepFixture(t)
	depImportPath := "github.com/tylergannon/go-gen-jsonschema/internal/builder/testfixtures/" + filepath.Base(depDir)

	targetDir := writeAsRefCollisionRootFixture(t, depImportPath)
	pkgs, err := syntax.Load(targetDir)
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	require.Empty(t, pkgs[0].Errors)
	scan, err := syntax.LoadPackage(pkgs[0])
	require.NoError(t, err)
	require.NotEmpty(t, scan.SchemaMethods)

	_, err = New(pkgs[0])
	require.ErrorContains(t, err, "AsRef definition name collision")
	require.ErrorContains(t, err, `"Shared"`)
}

// writeAsRefCollisionDepFixture writes a small dependency package, with no
// jsonschema build-tag constraints, exposing a "Shared" type with a Schema()
// method. It exists purely so the root fixture can register it as a second,
// distinct AsRef()'d type that happens to share its bare name with a locally
// declared "Shared" type.
func writeAsRefCollisionDepFixture(t *testing.T) string {
	t.Helper()

	cwd, err := os.Getwd()
	require.NoError(t, err)
	depDir, err := os.MkdirTemp(filepath.Join(cwd, "testfixtures"), "asref_collision_dep_")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(depDir))
	})

	source := `package ` + filepath.Base(depDir) + `

import "encoding/json"

type Shared struct {
	Name string ` + "`json:\"name\"`" + `
}

func (Shared) Schema() json.RawMessage { panic("not implemented") }
`
	require.NoError(t, os.WriteFile(filepath.Join(depDir, "shared.go"), []byte(source), 0o644))
	return depDir
}

func writeAsRefCollisionRootFixture(t *testing.T, depImportPath string) string {
	t.Helper()

	cwd, err := os.Getwd()
	require.NoError(t, err)
	targetDir, err := os.MkdirTemp(filepath.Join(cwd, "testfixtures"), "asref_collision_root_")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(targetDir))
	})

	source := `//go:build jsonschema

package fixture

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	dep "` + depImportPath + `"
)

type Shared struct {
	Name string ` + "`json:\"name\"`" + `
}

func (Shared) Schema() json.RawMessage { panic("not implemented") }

type Container struct {
	Local  Shared     ` + "`json:\"local\"`" + `
	Remote dep.Shared ` + "`json:\"remote\"`" + `
}

func (Container) Schema() json.RawMessage { panic("not implemented") }

var (
	_ = jsonschema.NewJSONSchemaMethod(Shared.Schema, jsonschema.AsRef())
	_ = jsonschema.NewJSONSchemaMethod(dep.Shared.Schema, jsonschema.AsRef())
	_ = jsonschema.NewJSONSchemaMethod(Container.Schema)
)
`
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "schema.go"), []byte(source), 0o644))
	return targetDir
}
