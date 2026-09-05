package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylergannon/go-gen-jsonschema/internal/syntax"
)

// writeFluentFixture writes a single-file package to a fresh temp directory
// under testfixtures/ and builds it in-process (no go.mod/go generate),
// mirroring writeInlineInterfaceFixture's pattern.
func writeFluentFixture(t *testing.T, source string) SchemaBuilder {
	t.Helper()

	cwd, err := os.Getwd()
	require.NoError(t, err)
	targetDir, err := os.MkdirTemp(filepath.Join(cwd, "testfixtures"), "fluent_")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(targetDir))
	})

	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "schema.go"), []byte(source), 0o644))

	pkgs, err := syntax.Load(targetDir)
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	require.Empty(t, pkgs[0].Errors)

	builder, err := New(pkgs[0])
	require.NoError(t, err)
	return builder
}

// jsonFor renders a built type's JSON schema through the same
// marshalSchemaHardlines path used for generated .json/.json.tmpl files, so
// legacy and fluent registrations of the "same" type can be compared for
// byte-for-byte parity.
func jsonFor(t *testing.T, b SchemaBuilder, typeName string) string {
	t.Helper()
	schema, ok := b.schemas.Get(b.Scan.Pkg.PkgPath, typeName)
	require.True(t, ok, "no schema recorded for %s", typeName)
	out, err := marshalSchemaHardlines(schema)
	require.NoError(t, err)
	return string(out)
}

const fluentProviderFixture = `//go:build jsonschema

package fixture

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

type Example struct {
	A string ` + "`json:\"a\"`" + `
	B int    ` + "`json:\"b\"`" + `
	C bool   ` + "`json:\"c\"`" + `
}

func (Example) Schema() json.RawMessage { panic("not implemented") }
func (Example) ASchema() json.Marshaler {
	return json.RawMessage(` + "`{\"type\":\"string\",\"description\":\"A\"}`" + `)
}
func (Example) BSchema(_ int) json.Marshaler {
	return json.RawMessage(` + "`{\"type\":\"integer\",\"description\":\"B\"}`" + `)
}
func BoolSchemaFunc(_ bool) json.Marshaler {
	return json.RawMessage(` + "`{\"type\":\"boolean\",\"description\":\"C\"}`" + `)
}

var _ = %s
`

const legacyProviderRegistration = `jsonschema.NewJSONSchemaMethod(
	Example.Schema,
	jsonschema.WithStructAccessorMethod(Example{}.A, (Example).ASchema),
	jsonschema.WithStructFunctionMethod(Example{}.B, (Example).BSchema),
	jsonschema.WithFunction(Example{}.C, BoolSchemaFunc),
	jsonschema.WithRenderProviders(),
)`

const fluentProviderRegistration = `jsonschema.Declare(Example.Schema).
	Accessor(Example{}.A, Example.ASchema).
	Method(Example{}.B, Example.BSchema).
	Function(Example{}.C, BoolSchemaFunc).
	RenderProviders()`

// TestFluentProviderParityWithLegacy proves that Accessor/Method/Function/
// RenderProviders fluent chaining produces the exact same rendered schema as
// the equivalent NewJSONSchemaMethod + WithXxx registration, through the
// same builder path used for real generation.
func TestFluentProviderParityWithLegacy(t *testing.T) {
	t.Parallel()

	legacy := writeFluentFixture(t, fmt.Sprintf(fluentProviderFixture, legacyProviderRegistration))
	fluent := writeFluentFixture(t, fmt.Sprintf(fluentProviderFixture, fluentProviderRegistration))

	require.Equal(t, jsonFor(t, legacy, "Example"), jsonFor(t, fluent, "Example"))
}

const fluentPointerProviderFixture = `//go:build jsonschema

package fixture

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

type Example struct {
	A string ` + "`json:\"a\"`" + `
	B int    ` + "`json:\"b\"`" + `
}

func (*Example) Schema() json.RawMessage { panic("not implemented") }
func (*Example) ASchema() json.Marshaler {
	return json.RawMessage(` + "`{\"type\":\"string\",\"description\":\"A\"}`" + `)
}
func (*Example) BSchema(_ int) json.Marshaler {
	return json.RawMessage(` + "`{\"type\":\"integer\",\"description\":\"B\"}`" + `)
}

var _ = %s
`

const legacyPointerProviderRegistration = `jsonschema.NewJSONSchemaMethod(
	(*Example).Schema,
	jsonschema.WithStructAccessorMethod(Example{}.A, (*Example).ASchema),
	jsonschema.WithStructFunctionMethod(Example{}.B, (*Example).BSchema),
)`

const fluentPointerProviderRegistration = `jsonschema.Declare((*Example).Schema).
	Accessor(Example{}.A, (*Example).ASchema).
	Method(Example{}.B, (*Example).BSchema)`

// TestFluentPointerRootProviderParityWithLegacy proves that a pointer-root
// fluent chain (Declare((*T).Schema).Accessor/.Method with pointer method
// expressions) actually retains its providers, matching the equivalent
// pointer-receiver legacy registration byte-for-byte. This reproduces the
// issue #73 review finding: providerRef previously rejected the *dst.StarExpr
// inside "(*Example).ASchema" and silently dropped the provider option.
func TestFluentPointerRootProviderParityWithLegacy(t *testing.T) {
	t.Parallel()

	legacy := writeFluentFixture(t, fmt.Sprintf(fluentPointerProviderFixture, legacyPointerProviderRegistration))
	fluent := writeFluentFixture(t, fmt.Sprintf(fluentPointerProviderFixture, fluentPointerProviderRegistration))

	legacyJSON := jsonFor(t, legacy, "Example")
	require.Equal(t, legacyJSON, jsonFor(t, fluent, "Example"))
	require.Contains(t, legacyJSON, `{{.a}}`)
	require.Contains(t, legacyJSON, `{{.b}}`)
}

const fluentEnumFixture = `//go:build jsonschema

package fixture

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

type Paint string

const (
	Red   Paint = "red"
	Green Paint = "green"
)

func (p Paint) String() string { return string(p) }

type Widget struct {
	Direct  Paint ` + "`json:\"direct\"`" + `
	ViaStringer Paint ` + "`json:\"viaStringer\"`" + `
}

func (Widget) Schema() json.RawMessage { panic("not implemented") }

var _ = %s
`

const legacyEnumRegistration = `jsonschema.NewJSONSchemaMethod(
	Widget.Schema,
	jsonschema.WithEnum(Widget{}.Direct),
	jsonschema.WithStringerEnum(Widget{}.ViaStringer),
)`

const fluentEnumRegistration = `jsonschema.Declare(Widget.Schema).
	Enum(Widget{}.Direct).
	StringerEnum(Widget{}.ViaStringer)`

// TestFluentEnumParityWithLegacy proves that .Enum/.StringerEnum chaining
// produces the exact same rendered schema as WithEnum/WithStringerEnum.
func TestFluentEnumParityWithLegacy(t *testing.T) {
	t.Parallel()

	legacy := writeFluentFixture(t, fmt.Sprintf(fluentEnumFixture, legacyEnumRegistration))
	fluent := writeFluentFixture(t, fmt.Sprintf(fluentEnumFixture, fluentEnumRegistration))

	require.Equal(t, jsonFor(t, legacy, "Widget"), jsonFor(t, fluent, "Widget"))
}

const fluentRefFixture = `//go:build jsonschema

package fixture

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

type Shared struct {
	Name string ` + "`json:\"name\"`" + `
}

func (Shared) Schema() json.RawMessage { panic("not implemented") }

type Owner struct {
	Value Shared ` + "`json:\"value\"`" + `
}

func (Owner) Schema() json.RawMessage { panic("not implemented") }

var _ = %s
var _ = jsonschema.NewJSONSchemaMethod(Owner.Schema)
`

const legacyRefRegistration = `jsonschema.NewJSONSchemaMethod(Shared.Schema, jsonschema.AsRef())`
const fluentRefRegistration = `jsonschema.Declare(Shared.Schema).Ref()`

// TestFluentRefParityWithLegacy proves that .Ref() produces the same
// "$ref"-based owner schema as AsRef().
func TestFluentRefParityWithLegacy(t *testing.T) {
	t.Parallel()

	legacy := writeFluentFixture(t, fmt.Sprintf(fluentRefFixture, legacyRefRegistration))
	fluent := writeFluentFixture(t, fmt.Sprintf(fluentRefFixture, fluentRefRegistration))

	require.Equal(t, jsonFor(t, legacy, "Owner"), jsonFor(t, fluent, "Owner"))
	require.Contains(t, jsonFor(t, legacy, "Owner"), `"$ref"`)
}

const fluentInterfaceFixture = `//go:build jsonschema

package fixture

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

type Value interface{ value() }

type First struct {
	Name string ` + "`json:\"name\"`" + `
}

func (First) value() {}

type Second struct {
	Count int ` + "`json:\"count\"`" + `
}

func (Second) value() {}

type Owner struct {
	Value Value ` + "`json:\"value\"`" + `
}

func (Owner) Schema() json.RawMessage { panic("not implemented") }

var _ = %s
`

const legacyInterfaceRegistration = `jsonschema.NewJSONSchemaMethod(
	Owner.Schema,
	jsonschema.WithInterface(
		Owner{}.Value,
		jsonschema.Discriminator("kind"),
		jsonschema.Impl("first", First{}),
		jsonschema.Impl("second", Second{}),
	),
)`

const fluentInterfaceRegistration = `jsonschema.Declare(Owner.Schema).
	Interface(
		Owner{}.Value,
		jsonschema.Discriminator("kind"),
		jsonschema.Impl("first", First{}),
		jsonschema.Impl("second", Second{}),
	)`

// TestFluentInterfaceParityWithLegacy proves that a sealed-interface fluent
// declaration (.Interface with inline Discriminator/Impl options) produces
// the same schema as the equivalent inline WithInterface(...) registration.
func TestFluentInterfaceParityWithLegacy(t *testing.T) {
	t.Parallel()

	legacy := writeFluentFixture(t, fmt.Sprintf(fluentInterfaceFixture, legacyInterfaceRegistration))
	fluent := writeFluentFixture(t, fmt.Sprintf(fluentInterfaceFixture, fluentInterfaceRegistration))

	require.Equal(t, jsonFor(t, legacy, "Owner"), jsonFor(t, fluent, "Owner"))
}

// TestFluentInterfaceRegistrationDiagnosticsParity proves that a fluent
// .Interface(...) declaration surfaces the exact same builder-level
// diagnostic as the equivalent inline WithInterface(...) call when an Impl
// type doesn't satisfy the interface, reusing the same error path rather
// than a fluent-specific one.
func TestFluentInterfaceRegistrationDiagnosticsParity(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	require.NoError(t, err)
	targetDir, err := os.MkdirTemp(filepath.Join(cwd, "testfixtures"), "fluent_interface_diag_")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(targetDir))
	})

	source := `//go:build jsonschema

package fixture

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

type Value interface{ value() }

type First struct { Name string ` + "`json:\"name\"`" + ` }
func (First) value() {}

type Stranger struct { Enabled bool ` + "`json:\"enabled\"`" + ` }

type Owner struct { Value Value ` + "`json:\"value\"`" + ` }
func (Owner) Schema() json.RawMessage { panic("not implemented") }

var _ = jsonschema.Declare(Owner.Schema).
	Interface(Owner{}.Value, jsonschema.Impl("stranger", Stranger{}))
`
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "schema.go"), []byte(source), 0o644))

	pkgs, err := syntax.Load(targetDir)
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	require.Empty(t, pkgs[0].Errors)

	_, err = New(pkgs[0])
	require.ErrorContains(t, err, "does not implement Value")
}

// TestFluentAccessorRejectsFreeFunctionProvider proves the issue #73 review
// finding is fixed: passing a free function shaped like func(T)
// json.Marshaler to .Accessor (where Go's type system accepts it identically
// to the intended receiver method expression) is now a source-positioned
// scanner error, not a silently-produced FieldProvider that the code-gen
// template has no branch for and that panics at runtime.
func TestFluentAccessorRejectsFreeFunctionProvider(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	require.NoError(t, err)
	targetDir, err := os.MkdirTemp(filepath.Join(cwd, "testfixtures"), "fluent_accessor_shape_")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(targetDir))
	})

	source := `//go:build jsonschema

package fixture

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

type Example struct {
	A string ` + "`json:\"a\"`" + `
}

func (Example) Schema() json.RawMessage { panic("not implemented") }

func freeAccessorSchema(Example) json.Marshaler { panic("not implemented") }

var _ = jsonschema.Declare(Example.Schema).
	Accessor(Example{}.A, freeAccessorSchema).
	RenderProviders()
`
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "schema.go"), []byte(source), 0o644))

	pkgs, err := syntax.Load(targetDir)
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	require.Empty(t, pkgs[0].Errors)

	_, err = New(pkgs[0])
	require.ErrorContains(t, err, "jsonschema.Declare: .Accessor provider must be a Example method expression, not a free function")
	require.ErrorContains(t, err, "schema.go")
}

// TestFluentFunctionRejectsUnrelatedMethodExpressionProvider proves the
// second reachable path to the same review finding: providerRef reports
// matched=false (rather than isMethod=true) when a chain link's provider is
// a method expression whose receiver isn't recognized against this chain
// (e.g. a method on the field's own type, passed to .Function, which
// requires a free function). Go's type system accepts this because
// .Function's provider signature is func(F) json.Marshaler, where F is the
// field's own type rather than the Declare root - so the receiver mismatch
// can only be caught here, and must be a hard error rather than the silent
// option-drop the "matched=false" case previously fell through to.
func TestFluentFunctionRejectsUnrelatedMethodExpressionProvider(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	require.NoError(t, err)
	targetDir, err := os.MkdirTemp(filepath.Join(cwd, "testfixtures"), "fluent_function_shape_")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(targetDir))
	})

	source := `//go:build jsonschema

package fixture

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

type Passthrough struct{}

func (Passthrough) PassthroughSchema() json.Marshaler { panic("not implemented") }

type Example struct {
	H Passthrough ` + "`json:\"h\"`" + `
}

func (Example) Schema() json.RawMessage { panic("not implemented") }

var _ = jsonschema.Declare(Example.Schema).
	Function(Example{}.H, Passthrough.PassthroughSchema).
	RenderProviders()
`
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "schema.go"), []byte(source), 0o644))

	pkgs, err := syntax.Load(targetDir)
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	require.Empty(t, pkgs[0].Errors)

	_, err = New(pkgs[0])
	require.ErrorContains(t, err, "jsonschema.Declare: .Function provider is not a supported method expression or free function")
	require.ErrorContains(t, err, "schema.go")
}
