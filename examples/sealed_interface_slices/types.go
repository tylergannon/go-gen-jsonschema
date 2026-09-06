// Package sealed_interface_slices is the acceptance example for direct slices
// of registered interface unions. The session worklog records its contract.
package sealed_interface_slices

//go:generate go run ../../polytype/ --pretty

// Event is a sealed union: its unexported isEvent method closes membership
// to the same-package struct types that declare it, and the generator infers
// exactly those types as the wire variants.
type Event interface {
	isEvent()
}

// Created is a value implementation of Event.
type Created struct {
	Name string `json:"name"`
}

func (Created) isEvent() {}

// Deleted is a pointer implementation of Event.
type Deleted struct {
	ID string `json:"id"`
}

func (*Deleted) isEvent() {}

// Batch is the desired supported shape: one direct, one-dimensional slice of
// a registered interface union on a named struct field.
type Batch struct {
	Events []Event `json:"events"`
}
