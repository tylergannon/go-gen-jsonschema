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

// NewJSONSchemaBuilder registers a function as being a stub that should be
// implemented with a proper json schema and, as needed, unmarshaler functionality.
func NewJSONSchemaBuilder[T any](SchemaFunction) SchemaMarker {
	return SchemaMarker{}
}

// NewJSONSchemaMethod registers a struct method as a stub that will be implemented
// with a proper json schema and, as needed, unmarshaler functionality.
func NewJSONSchemaMethod[T any](SchemaMethod[T]) SchemaMarker {
	return SchemaMarker{}
}

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
