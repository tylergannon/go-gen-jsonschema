package jsonschema

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	errOptionalAbsent = errors.New("cannot marshal an absent Optional value")
	errPresentNull    = errors.New("present value marshaled as JSON null")
)

// Optional represents an object property that may be absent. The zero value is
// absent; a present value may contain T's zero value but may not encode as null.
// Containing struct fields must use json:",omitzero" so absent values are
// omitted before MarshalJSON is called.
type Optional[T any] struct {
	Present bool
	Value   T
}

// IsZero reports whether the property is absent.
func (o Optional[T]) IsZero() bool { return !o.Present }

// MarshalJSON encodes a present non-null value.
func (o Optional[T]) MarshalJSON() ([]byte, error) {
	if !o.Present {
		return nil, errOptionalAbsent
	}
	return marshalPresent(o.Value)
}

// UnmarshalJSON decodes a present non-null value without mutating the receiver
// when decoding fails.
func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	if isJSONNull(data) {
		return errors.New("Optional value cannot be JSON null")
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*o = Optional[T]{Present: true, Value: value}
	return nil
}

// Nullable represents a required object property whose value may be null. The
// zero value encodes as null; Present reports whether Value is non-null.
type Nullable[T any] struct {
	Present bool
	Value   T
}

// IsZero always reports false so json:",omitzero" cannot omit a required
// nullable property.
func (Nullable[T]) IsZero() bool { return false }

// MarshalJSON encodes null or a present non-null value.
func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if !n.Present {
		return []byte("null"), nil
	}
	return marshalPresent(n.Value)
}

// UnmarshalJSON decodes null or a present value without mutating the receiver
// when decoding fails.
func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	if isJSONNull(data) {
		var zero T
		*n = Nullable[T]{Value: zero}
		return nil
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*n = Nullable[T]{Present: true, Value: value}
	return nil
}

func marshalPresent[T any](value T) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if isJSONNull(data) {
		return nil, fmt.Errorf("%w: %T", errPresentNull, value)
	}
	return data, nil
}

func isJSONNull(data []byte) bool {
	return bytes.Equal(bytes.TrimSpace(data), []byte("null"))
}
