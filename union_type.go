package jsonschema

import (
	"encoding/json"
)

type (
	EnumType     struct{}
	SchemaMarker struct{}

	InterfaceMarker struct{}

	SchemaFunction func() json.RawMessage

	SchemaMethod[T any] func(T) json.RawMessage
)

// WithRenderProviders requests generation of RenderedSchema() and provider execution at runtime.
func WithRenderProviders() SchemaMethodOption { return SchemaMethodOptionObj{} }

// NewJSONSchemaBuilder registers a function as being a stub that should be
// implemented with a proper json schema and, as needed, unmarshaler functionality.
func NewJSONSchemaBuilder[T any](SchemaFunction) SchemaMarker {
	return SchemaMarker{}
}

// NewJSONSchemaBuilderFor registers a zero-arg builder function for the given
// example instance value (e.g., TypeName{}), allowing type inference without
// generics.
func NewJSONSchemaBuilderFor(_ any, _ SchemaFunction, _ ...SchemaMethodOption) SchemaMarker {
	return SchemaMarker{}
}

type SchemaMethodOption interface {
	implementsSchemaMethodOption()
}

type exampleStruct struct {
	Field1 string
	Field2 int
	Field3 bool
}

func buildBoolSchema(val bool) json.Marshaler {
	return json.RawMessage(`{"type": "boolean"}`)
}

func (exampleStruct) field1Schema() json.Marshaler {
	return json.RawMessage(`{"type": "string"}`)
}

func (exampleStruct) field2Schema(int) json.Marshaler {
	return json.RawMessage(`{"type": "integer"}`)
}

func (exampleStruct) JSONSchema() json.RawMessage {
	panic("not implemented")
}

func WithFunction[T any](val T, f func(T) json.Marshaler) SchemaMethodOption {
	return SchemaMethodOptionObj{}
}

func WithStructFunctionMethod[T, U any](val U, f func(T, U) json.Marshaler) SchemaMethodOption {
	return SchemaMethodOptionObj{}
}

func WithStructAccessorMethod[T, U any](val T, f func(U) json.Marshaler) SchemaMethodOption {
	return SchemaMethodOptionObj{}
}

type SchemaMethodOptionObj struct{}

func (SchemaMethodOptionObj) implementsSchemaMethodOption() {}

// Interface options (v1) - stubs for scanning/type-checking; parsed by scanner
func WithInterface[T any](field T) SchemaMethodOption { return SchemaMethodOptionObj{} }
func WithInterfaceImpls[T any](field T, impls ...any) SchemaMethodOption {
	return SchemaMethodOptionObj{}
}
func WithDiscriminator[T any](field T, name string) SchemaMethodOption {
	return SchemaMethodOptionObj{}
}

// Enum options (v1) - stubs for scanning/type-checking; parsed by scanner
type EnumMode int

const (
	EnumStrings EnumMode = iota + 1
)

func WithEnum[T any](field T) SchemaMethodOption                  { return SchemaMethodOptionObj{} }
func WithEnumMode(mode EnumMode) SchemaMethodOption               { return SchemaMethodOptionObj{} }
func WithEnumName[T any](value T, name string) SchemaMethodOption { return SchemaMethodOptionObj{} }

// NewJSONSchemaMethod registers a struct method as a stub that will be implemented
// with a proper json schema and, as needed, unmarshaler functionality.
func NewJSONSchemaMethod[T any](SchemaMethod[T], ...SchemaMethodOption) SchemaMarker {
	return SchemaMarker{}
}

// NewJSONSchemaFunc registers a free function that takes the receiver as its
// sole parameter as a schema entrypoint. It is equivalent to NewJSONSchemaMethod.
func NewJSONSchemaFunc[T any](f SchemaMethod[T], _ ...SchemaMethodOption) SchemaMarker { // options parsed by scanner only for now
	_ = f
	return SchemaMarker{}
}

var _ SchemaMarker = NewJSONSchemaMethod(
	exampleStruct.JSONSchema,
	WithStructAccessorMethod(exampleStruct{}.Field1, exampleStruct.field1Schema),
	WithStructFunctionMethod(exampleStruct{}.Field2, exampleStruct.field2Schema),
	WithFunction(exampleStruct{}.Field3, buildBoolSchema),
)

// NewInterfaceImpl marks the arguments as possible implementations for the
// interface type given in the type argument.
//  1. If called in the same package as the interface itself, then all global
//     instances can be replaced.
//  2. If called somewhere else, only applies to the local package.
func NewInterfaceImpl[T any](...T) InterfaceMarker {
	return InterfaceMarker{}
}

// NewEnumType denotes that the type argument should be an enum.
// If called in the same package where the type is declared, then
// it applies globally.
// In all cases, the const values MUST be declared in the same
// package as the call to NewEnumType.
//
// For now, only string types are supported.
func NewEnumType[T ~string]() EnumType {
	return EnumType{}
}
