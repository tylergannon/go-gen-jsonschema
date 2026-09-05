package consumer

import (
	"fmt"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

type Event interface{ event() }

type Created struct {
	Name string `json:"name"`
}

func (Created) event() {}

type Deleted struct {
	ID string `json:"id"`
}

func (*Deleted) event() {}

type State int

const (
	StateUnknown State = iota
	StateNew
	StateDone
)

func (state State) String() string { return fmt.Sprintf("human-state-%d", state) }

type Metadata struct {
	Visible string `json:"visible"`
	Hidden  string `json:"-"`
}

type Envelope struct {
	Primary       Event                      `json:"primary"`
	Alternate     Event                      `json:"alternate"`
	Maybe         jsonschema.Optional[Event] `json:"maybe,omitzero"`
	Events        []Event                    `json:"events"`
	StringState   State                      `json:"string_state"`
	NumericState  State                      `json:"numeric_state"`
	OptionalState jsonschema.Optional[State] `json:"optional_state,omitzero"`
	NullableState jsonschema.Nullable[State] `json:"nullable_state"`
	NullState     jsonschema.Nullable[State] `json:"null_state"`
	Label         string                     `json:"ordinary-name"`
	Meta          Metadata                   `json:"meta"`
	Ignored       string                     `json:"-"`
}
