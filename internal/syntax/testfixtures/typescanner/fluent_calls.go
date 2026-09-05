//go:build jsonschema

package typescanner

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	alias "github.com/tylergannon/go-gen-jsonschema/internal/syntax/testfixtures/typescanner/scannersubpkg"
)

type FluentStruct struct {
	A string
	B int
	C bool
	E NiceEnumType
	F NiceEnumType
	G MarkerInterface
}

func (FluentStruct) Schema() json.RawMessage        { panic("not implemented") }
func (FluentStruct) ASchema() json.Marshaler        { panic("not implemented") }
func (FluentStruct) BSchema(int) json.Marshaler     { panic("not implemented") }
func FluentBoolSchema(bool) json.Marshaler          { panic("not implemented") }
func FluentFreeSchema(FluentStruct) json.RawMessage { panic("not implemented") }

func (t *FluentStruct) PtrSchema() json.RawMessage    { panic("not implemented") }
func (t *FluentStruct) PtrASchema() json.Marshaler    { panic("not implemented") }
func (t *FluentStruct) PtrBSchema(int) json.Marshaler { panic("not implemented") }

// One fluent chain exercising every supported chain method.
var _ = jsonschema.Declare(FluentStruct.Schema).
	Accessor(FluentStruct{}.A, FluentStruct.ASchema).
	Method(FluentStruct{}.B, FluentStruct.BSchema).
	Function(FluentStruct{}.C, FluentBoolSchema).
	Enum(FluentStruct{}.E).
	StringerEnum(FluentStruct{}.F).
	Interface(
		FluentStruct{}.G,
		jsonschema.Discriminator("kind"),
		jsonschema.Impl("one", Type001{}),
	).
	Ref().
	RenderProviders()

// A pointer-root fluent chain, proving providerRef recognizes the
// (*FluentStruct).Method form alongside the value-receiver form above.
var _ = jsonschema.Declare((*FluentStruct).PtrSchema).
	Accessor(FluentStruct{}.A, (*FluentStruct).PtrASchema).
	Method(FluentStruct{}.B, (*FluentStruct).PtrBSchema)

// A free-function root, with no chained options.
var _ = jsonschema.Declare(FluentFreeSchema)

// A bare Declare with no chain at all.
var _ = jsonschema.Declare(FluentStruct.Schema)

// The same alias-imported package used elsewhere in this fixture, proving
// import aliases resolve identically through the fluent chain-walk (which
// reuses the same IdentifyFunc/GetPackageForPrefix machinery).
var _ = jsonschema.Declare(alias.TypeForSchemaMethod.Schema).
	Ref()
