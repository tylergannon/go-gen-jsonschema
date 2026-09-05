package polytype

import "encoding/json"

// Declaration is the typed fluent builder for registering a schema
// entrypoint. Like every other marker type in this package, it exists only
// inside `//go:build jsonschema` files: it performs no work at runtime, is
// never called, and exists solely to type-check and to give the scanner a
// recognizable AST shape.
type Declaration[T any] struct{}

// Declare registers fn as the schema entrypoint for T. fn may be either a
// method expression (e.g. Example.Schema) or a free function taking T as its
// sole parameter (e.g. BuildExampleSchema); both forms infer T from fn's
// signature. Chain the returned *Declaration[T] with Accessor, Method,
// Function, Enum, StringerEnum, Ref, RenderProviders, and Interface to add
// options, matching the equivalent WithXxx options on NewJSONSchemaMethod.
func Declare[T any](fn func(T) json.RawMessage) *Declaration[T] {
	_ = fn
	return &Declaration[T]{}
}

// Accessor registers a provider for field that is a struct method taking
// only the receiver T (equivalent to WithStructAccessorMethod).
func (d *Declaration[T]) Accessor(field any, provider func(T) json.Marshaler) *Declaration[T] {
	return d
}

// Method registers a provider for field that is a struct method also taking
// the field's own value F (equivalent to WithStructFunctionMethod). field and
// provider must agree on F: passing a field of one type alongside a provider
// expecting another fails to compile.
func (d *Declaration[T]) Method[F any](field F, provider func(T, F) json.Marshaler) *Declaration[T] {
	return d
}

// Function registers a provider for field that is a free function taking
// only the field's own value F (equivalent to WithFunction). field and
// provider must agree on F: passing a field of one type alongside a provider
// expecting another fails to compile.
func (d *Declaration[T]) Function[F any](field F, provider func(F) json.Marshaler) *Declaration[T] {
	return d
}

// Enum marks field as an enum whose values are compared directly
// (equivalent to WithEnum).
func (d *Declaration[T]) Enum(field any) *Declaration[T] {
	return d
}

// StringerEnum marks field as an enum whose values are compared via
// fmt.Stringer (equivalent to WithStringerEnum).
func (d *Declaration[T]) StringerEnum(field any) *Declaration[T] {
	return d
}

// Ref requests that, wherever T is referenced from another registered
// schema, it be rendered as a "$ref" into that schema's "$defs" instead of
// being inlined (equivalent to AsRef).
func (d *Declaration[T]) Ref() *Declaration[T] {
	return d
}

// RenderProviders requests generation of RenderedSchema() and provider
// execution at runtime (equivalent to WithRenderProviders).
func (d *Declaration[T]) RenderProviders() *Declaration[T] {
	return d
}

// Interface marks field as a sealed interface, configured via Discriminator
// and Impl options (equivalent to WithInterface).
func (d *Declaration[T]) Interface(field any, options ...InterfaceOption) *Declaration[T] {
	return d
}
