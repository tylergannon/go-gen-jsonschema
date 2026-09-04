//go:build jsonschema

package inspection_nested

import (
	"encoding/json"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (Child) Schema() json.RawMessage          { panic("not implemented") }
func (Parent) Schema() json.RawMessage         { panic("not implemented") }
func (HookModel) Schema() json.RawMessage      { panic("not implemented") }
func (WireMismatch) Schema() json.RawMessage   { panic("not implemented") }
func (ProviderModel) Schema() json.RawMessage  { panic("not implemented") }
func (ExternalModel) Schema() json.RawMessage  { panic("not implemented") }
func (BadUnion) Schema() json.RawMessage       { panic("not implemented") }
func (StubOnly) Schema() json.RawMessage       { panic("not implemented") }
func (RemoteSameName) Schema() json.RawMessage { panic("not implemented") }

func (StubOnly) MarshalJSON() ([]byte, error) { panic("generation stub") }

var (
	_ = jsonschema.NewJSONSchemaMethod(
		Child.Schema,
		jsonschema.WithInterface(Child{}.Event, jsonschema.Impl("created", Created{})),
		jsonschema.WithStringerEnum(Child{}.State),
	)
	_ = jsonschema.NewJSONSchemaMethod(Parent.Schema)
	_ = jsonschema.NewJSONSchemaMethod(HookModel.Schema)
	_ = jsonschema.NewJSONSchemaMethod(WireMismatch.Schema)
	_ = jsonschema.NewJSONSchemaMethod(
		ProviderModel.Schema,
		jsonschema.WithFunction(ProviderModel{}.Value, ProviderSchema),
	)
	_ = jsonschema.NewJSONSchemaMethod(ExternalModel.Schema)
	_ = jsonschema.NewJSONSchemaMethod(
		BadUnion.Schema,
		jsonschema.WithInterface(BadUnion{}.Events, jsonschema.Impl("created", Created{})),
	)
	_ = jsonschema.NewJSONSchemaMethod(StubOnly.Schema)
	_ = jsonschema.NewJSONSchemaMethod(RemoteSameName.Schema)
)
