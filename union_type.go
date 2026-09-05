package polytype

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
//
// Deprecated: use Declare(T.Schema).RenderProviders() instead.
func WithRenderProviders() SchemaMethodOption { return SchemaMethodOptionObj{} }

// AsRef requests that, wherever this type is referenced from another
// registered schema, it be rendered as a "$ref" into that schema's "$defs"
// instead of being inlined.
//
// Deprecated: use Declare(T.Schema).Ref() instead.
func AsRef() SchemaMethodOption { return SchemaMethodOptionObj{} }

// NewJSONSchemaBuilder registers a function as being a stub that should be
// implemented with a proper json schema and, as needed, unmarshaler functionality.
func NewJSONSchemaBuilder[T any](SchemaFunction) SchemaMarker {
	return SchemaMarker{}
}

type SchemaMethodOption interface {
	implementsSchemaMethodOption()
}

// InterfaceOption configures a registered interface field.
type InterfaceOption interface {
	implementsInterfaceOption()
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

// Deprecated: use Declare(T.Schema).Function(field, fn) instead.
func WithFunction[T any](val T, f func(T) json.Marshaler) SchemaMethodOption {
	return SchemaMethodOptionObj{}
}

// Deprecated: use Declare(T.Schema).Method(field, T.method) instead.
func WithStructFunctionMethod[T, U any](val U, f func(T, U) json.Marshaler) SchemaMethodOption {
	return SchemaMethodOptionObj{}
}

// Deprecated: use Declare(T.Schema).Accessor(field, T.method) instead.
func WithStructAccessorMethod[T, U any](val T, f func(U) json.Marshaler) SchemaMethodOption {
	return SchemaMethodOptionObj{}
}

type SchemaMethodOptionObj struct{}

func (SchemaMethodOptionObj) implementsSchemaMethodOption() {}

type InterfaceOptionObj struct{}

func (InterfaceOptionObj) implementsInterfaceOption() {}

// Interface options (v1) - stubs for scanning/type-checking; parsed by scanner
//
// Deprecated: use Declare(T.Schema).Interface(field, options...) instead.
func WithInterface[T any](field T, options ...InterfaceOption) SchemaMethodOption {
	return SchemaMethodOptionObj{}
}

// Discriminator sets the JSON property used to distinguish interface cases.
func Discriminator(name string) InterfaceOption { return InterfaceOptionObj{} }

// Impl registers an interface implementation with its stable wire value.
func Impl[T any](value string, impl T) InterfaceOption { return InterfaceOptionObj{} }

// Deprecated: use Declare(T.Schema).Interface(field, Impl(value, impl), ...) instead.
func WithInterfaceImpls[T any](field T, impls ...any) SchemaMethodOption {
	return SchemaMethodOptionObj{}
}

// Deprecated: use Declare(T.Schema).Interface(field, Discriminator(name), ...) instead.
func WithDiscriminator[T any](field T, name string) SchemaMethodOption {
	return SchemaMethodOptionObj{}
}

// Enum options (v1) - stubs for scanning/type-checking; parsed by scanner
//
// Deprecated: use Declare(T.Schema).Enum(field) instead.
func WithEnum[T any](field T) SchemaMethodOption { return SchemaMethodOptionObj{} }

// Deprecated: use Declare(T.Schema).StringerEnum(field) instead.
func WithStringerEnum[T any](field T) SchemaMethodOption { return SchemaMethodOptionObj{} }

// NewJSONSchemaMethod registers a struct method as a stub that will be implemented
// with a proper json schema and, as needed, unmarshaler functionality.
//
// Deprecated: use Declare(T.Schema) instead. For example,
// NewJSONSchemaMethod(Task.Schema, WithEnum(Task{}.Status)) becomes
// Declare(Task.Schema).Enum(Task{}.Status).
func NewJSONSchemaMethod[T any](SchemaMethod[T], ...SchemaMethodOption) SchemaMarker {
	return SchemaMarker{}
}

// NewJSONSchemaFunc registers a free function that takes the receiver as its
// sole parameter as a schema entrypoint. It is equivalent to NewJSONSchemaMethod.
//
// Deprecated: use Declare(fn) with a free function instead.
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
//
// Deprecated: use Declare(T.Schema).Interface(field, Impl(value, impl), ...)
// on the field referencing the interface instead.
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
//
// Deprecated: use Declare(T.Schema).Enum(field) on the field referencing the
// enum type for direct field registrations. NewEnumType has no fluent
// replacement and must be retained when the enum type is shared across more
// than one struct field.
func NewEnumType[T ~string]() EnumType {
	return EnumType{}
}
