package builder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylergannon/go-gen-jsonschema/internal/syntax"
	"github.com/tylergannon/go-gen-jsonschema/internal/testutils"
)

func TestOwnerCodecRejectsProductionJSONMethodBeforeWriting(t *testing.T) {
	for _, method := range []string{"MarshalJSON", "UnmarshalJSON"} {
		t.Run(method, func(t *testing.T) {
			targetDir := writeOwnerCollisionFixture(t, "")
			declaration := `func (*Owner) UnmarshalJSON([]byte) error { return nil }`
			if method == "MarshalJSON" {
				declaration = `func (Owner) MarshalJSON() ([]byte, error) { return []byte("{}"), nil }`
			}
			require.NoError(t, os.WriteFile(filepath.Join(targetDir, "types.go"), []byte(ownerCollisionTypes+declaration), 0o644))

			err := Run(BuilderArgs{TargetDir: targetDir})
			require.ErrorContains(t, err, "handwritten production "+method)
			assertOwnerCollisionSentinels(t, targetDir)
		})
	}
}

func TestOwnerCodecAllowsGenerationOnlyDeclarationStub(t *testing.T) {
	targetDir := writeOwnerCollisionFixture(t, `func (Owner) MarshalJSON() ([]byte, error) { panic("not implemented") }`)
	pkgs, err := syntax.Load(targetDir)
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	_, err = New(pkgs[0])
	require.NoError(t, err)
}

func TestOwnerCodecRejectsPromotedProductionJSONMethod(t *testing.T) {
	targetDir := writeOwnerCollisionFixture(t, "")
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "types.go"), []byte(`package fixture

type Value interface{ value() }
type First struct{}
func (First) value() {}

type Hook struct{}
func (Hook) MarshalJSON() ([]byte, error) { return []byte("{}"), nil }

type Owner struct {
	Hook
	Value Value `+"`json:\"value\"`"+`
}
`), 0o644))
	err := Run(BuilderArgs{TargetDir: targetDir})
	require.ErrorContains(t, err, "handwritten production MarshalJSON already declared or promoted")
	assertOwnerCollisionSentinels(t, targetDir)
}

func TestOwnerCodecRejectsPromotedGeneratedOwner(t *testing.T) {
	targetDir := writeOwnerCollisionFixture(t, "")
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "types.go"), []byte(`package fixture

type Value interface{ value() }
type First struct{}
func (First) value() {}

type Embedded struct {
	Value Value `+"`json:\"value\"`"+`
}
type Owner struct { Embedded }
`), 0o644))
	schemaPath := filepath.Join(targetDir, "schema.go")
	schema, err := os.ReadFile(schemaPath)
	require.NoError(t, err)
	schema = append(schema, []byte(`
func (Embedded) Schema() json.RawMessage { panic("not implemented") }
var _ = jsonschema.NewJSONSchemaMethod(Embedded.Schema)
`)...)
	require.NoError(t, os.WriteFile(schemaPath, schema, 0o644))
	err = Run(BuilderArgs{TargetDir: targetDir})
	require.ErrorContains(t, err, "embedded type Embedded also requires generated owner codecs")
	assertOwnerCollisionSentinels(t, targetDir)
}

func TestOwnerCodecRejectsForeignEmbeddedGeneratedOwnerBeforeWriting(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	targetDir, err := os.MkdirTemp(filepath.Join(cwd, "testfixtures"), "foreign_embedded_owner_")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.RemoveAll(targetDir)) })
	baseImport := "github.com/tylergannon/go-gen-jsonschema/internal/builder/testfixtures/" + filepath.Base(targetDir)
	depDir := filepath.Join(targetDir, "dep")
	require.NoError(t, os.MkdirAll(depDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(depDir, "types.go"), []byte(`package dep

type Value interface{ value() }
type First struct{}
func (First) value() {}
type Embedded struct { Value Value `+"`json:\"value\"`"+` }
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(depDir, "schema.go"), []byte(`//go:build jsonschema

package dep

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Embedded) Schema() json.RawMessage { panic("not implemented") }
var _ = jsonschema.NewJSONSchemaMethod(Embedded.Schema)
var _ = jsonschema.NewInterfaceImpl[Value](First{})
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(depDir, "jsonschema_gen.go"), []byte(`//go:build !jsonschema

package dep

func (Embedded) MarshalJSON() ([]byte, error) { return []byte("{}"), nil }
func (*Embedded) UnmarshalJSON([]byte) error { return nil }
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "types.go"), []byte(`package fixture

import dep "`+baseImport+`/dep"

type Value interface{ value() }
type First struct{}
func (First) value() {}
type Owner struct {
	dep.Embedded
	Local Value `+"`json:\"local\"`"+`
}
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "schema.go"), []byte(`//go:build jsonschema

package fixture

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Owner) Schema() json.RawMessage { panic("not implemented") }
var _ = jsonschema.NewJSONSchemaMethod(Owner.Schema)
var _ = jsonschema.NewInterfaceImpl[Value](First{})
`), 0o644))
	writeOwnerCollisionSentinels(t, targetDir)

	err = Run(BuilderArgs{TargetDir: targetDir})
	require.ErrorContains(t, err, "foreign embedded type dep.Embedded has generated production JSON codecs")
	assertOwnerCollisionSentinels(t, targetDir)
}

func TestLegacyDuplicateDerivedDiscriminatorRejectedBeforeWriting(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	targetDir, err := os.MkdirTemp(filepath.Join(cwd, "testfixtures"), "duplicate_legacy_discriminator_")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.RemoveAll(targetDir)) })

	for _, pkg := range []string{"left", "right"} {
		require.NoError(t, os.MkdirAll(filepath.Join(targetDir, pkg), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(targetDir, pkg, "same.go"), []byte(`package `+pkg+`

type Same struct{}
func (Same) Marker() {}
`), 0o644))
	}
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "types.go"), []byte(`package fixture

type Value interface{ Marker() }
type Owner struct { Value Value `+"`json:\"value\"`"+` }
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "schema.go"), []byte(`//go:build jsonschema

package fixture

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"github.com/tylergannon/go-gen-jsonschema/internal/builder/testfixtures/`+filepath.Base(targetDir)+`/left"
	"github.com/tylergannon/go-gen-jsonschema/internal/builder/testfixtures/`+filepath.Base(targetDir)+`/right"
)

func (Owner) Schema() json.RawMessage { panic("not implemented") }
var _ = jsonschema.NewJSONSchemaMethod(Owner.Schema)
var _ = jsonschema.NewInterfaceImpl[Value](left.Same{}, right.Same{})
`), 0o644))
	writeOwnerCollisionSentinels(t, targetDir)

	err = Run(BuilderArgs{TargetDir: targetDir})
	require.ErrorContains(t, err, `duplicate discriminator value "Same"`)
	assertOwnerCollisionSentinels(t, targetDir)
}

func TestLegacyHelpersUseResolvedPackageIdentity(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	targetDir, err := os.MkdirTemp(filepath.Join(cwd, "testfixtures"), "legacy_helper_identity_")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.RemoveAll(targetDir)) })
	baseImport := "github.com/tylergannon/go-gen-jsonschema/internal/builder/testfixtures/" + filepath.Base(targetDir)

	packages := []struct {
		dir  string
		name string
	}{
		{dir: "left", name: "events"},
		{dir: "middle", name: "events1"},
		{dir: "right", name: "events"},
		{dir: "reserved", name: "json"},
	}
	for _, pkg := range packages {
		dir := filepath.Join(targetDir, pkg.dir)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "types.go"), []byte(`package `+pkg.name+`

type Event interface{ isEvent() }
type Created struct { Name string `+"`json:\"name\"`"+` }
func (Created) isEvent() {}
`), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "schema.go"), []byte(`//go:build jsonschema

package `+pkg.name+`

import jsonschema "github.com/tylergannon/go-gen-jsonschema"

var _ = jsonschema.NewInterfaceImpl[Event](Created{})
`), 0o644))
	}

	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "types.go"), []byte(`package fixture

import (
	left "`+baseImport+`/left"
	middle "`+baseImport+`/middle"
	right "`+baseImport+`/right"
	reserved "`+baseImport+`/reserved"
)

type Owner struct {
	Left left.Event `+"`json:\"left\"`"+`
	Middle middle.Event `+"`json:\"middle\"`"+`
	Right right.Event `+"`json:\"right\"`"+`
	Reserved reserved.Event `+"`json:\"reserved\"`"+`
}
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "schema.go"), []byte(`//go:build jsonschema

package fixture

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Owner) Schema() json.RawMessage { panic("not implemented") }
var _ = jsonschema.NewJSONSchemaMethod(Owner.Schema)
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "codec_test.go"), []byte(`package fixture

import (
	"encoding/json"
	"testing"
	left "`+baseImport+`/left"
	middle "`+baseImport+`/middle"
	right "`+baseImport+`/right"
	reserved "`+baseImport+`/reserved"
)

func TestDistinctSameNamedInterfacesRoundTrip(t *testing.T) {
	want := Owner{
		Left: left.Created{Name: "left"},
		Middle: middle.Created{Name: "middle"},
		Right: right.Created{Name: "right"},
		Reserved: reserved.Created{Name: "reserved"},
	}
	data, err := json.Marshal(want)
	if err != nil { t.Fatal(err) }
	var got Owner
	if err := json.Unmarshal(data, &got); err != nil { t.Fatal(err) }
	if value, ok := got.Left.(left.Created); !ok || value.Name != "left" { t.Fatalf("left = %#v", got.Left) }
	if value, ok := got.Middle.(middle.Created); !ok || value.Name != "middle" { t.Fatalf("middle = %#v", got.Middle) }
	if value, ok := got.Right.(right.Created); !ok || value.Name != "right" { t.Fatalf("right = %#v", got.Right) }
	if value, ok := got.Reserved.(reserved.Created); !ok || value.Name != "reserved" { t.Fatalf("reserved = %#v", got.Reserved) }
}
`), 0o644))

	require.NoError(t, Run(BuilderArgs{TargetDir: targetDir}))
	generated, err := os.ReadFile(filepath.Join(targetDir, "jsonschema_gen.go"))
	require.NoError(t, err)
	require.Equal(t, 2, strings.Count(string(generated), "func __jsonUnmarshal__events__Event__"))
	require.Equal(t, 1, strings.Count(string(generated), "func __jsonUnmarshal__events1__Event__"))
	require.Equal(t, 1, strings.Count(string(generated), "func __jsonUnmarshal__json__Event__"))
	require.Contains(t, string(generated), `events2 "`+baseImport+`/right"`)
	require.Contains(t, string(generated), `json1 "`+baseImport+`/reserved"`)
	exit, stdout, stderr, err := testutils.RunCommand("go", targetDir, "test", "./...")
	require.NoError(t, err)
	require.Equal(t, 0, exit, "stdout:\n%s\nstderr:\n%s", stdout, stderr)
}

const ownerCollisionTypes = `package fixture

type Value interface{ value() }
type First struct{}
func (First) value() {}
type Owner struct { Value Value ` + "`json:\"value\"`" + ` }
`

func writeOwnerCollisionFixture(t *testing.T, stub string) string {
	t.Helper()
	cwd, err := os.Getwd()
	require.NoError(t, err)
	targetDir, err := os.MkdirTemp(filepath.Join(cwd, "testfixtures"), "owner_codec_collision_")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.RemoveAll(targetDir)) })
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "types.go"), []byte(ownerCollisionTypes), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "schema.go"), []byte(`//go:build jsonschema

package fixture

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Owner) Schema() json.RawMessage { panic("not implemented") }
`+stub+`
var _ = jsonschema.NewJSONSchemaMethod(Owner.Schema)
var _ = jsonschema.NewInterfaceImpl[Value](First{})
`), 0o644))
	writeOwnerCollisionSentinels(t, targetDir)
	return targetDir
}

func writeOwnerCollisionSentinels(t *testing.T, targetDir string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(targetDir, "jsonschema"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "jsonschema", "Owner.json"), []byte("sentinel-schema"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "jsonschema_gen.go"), []byte("//go:build !jsonschema\n\npackage fixture\n\nvar sentinel = true\n"), 0o644))
}

func assertOwnerCollisionSentinels(t *testing.T, targetDir string) {
	t.Helper()
	schema, err := os.ReadFile(filepath.Join(targetDir, "jsonschema", "Owner.json"))
	require.NoError(t, err)
	require.Equal(t, "sentinel-schema", string(schema))
	generated, err := os.ReadFile(filepath.Join(targetDir, "jsonschema_gen.go"))
	require.NoError(t, err)
	require.Equal(t, "//go:build !jsonschema\n\npackage fixture\n\nvar sentinel = true\n", string(generated))
}
