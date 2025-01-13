package jsonschema

import "encoding/json"

type (
	SchemaMarker struct{}

	InterfaceImpl       struct{}
	SchemaFunction      func() (json.RawMessage, error)
	SchemaMethod[T any] func(T) (json.RawMessage, error)
)

// NewJSONSchemaBuilder registers a function as being a stub that should be
// implemented with a proper json schema and, as needed, unmarshaler functionality.
func NewJSONSchemaBuilder[T any](f func() (f SchemaFunction)) SchemaMarker {
	return SchemaMarker{}
}

// NewJSONSchemaMethod registers a struct method as a stub that will be implemented
// with a proper json schema and, as needed, unmarshaler functionality.
func NewJSONSchemaMethod[T any](f SchemaMethod[T]) SchemaMarker {
	return SchemaMarker{}
}

// NewInterfaceImpl marks the arguments as possible implementations for the
// interface type given in the type argument.
func NewInterfaceImpl[T any](implementations ...T) InterfaceImpl {
	return InterfaceImpl{}
}
