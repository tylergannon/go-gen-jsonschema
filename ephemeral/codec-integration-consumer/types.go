package boundary

import (
	"encoding/json"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

type Event interface{ event() }
type Created struct {
	Name string `json:"name"`
}

func (Created) event() {}

var Calls int

func (v Created) MarshalJSON() ([]byte, error) {
	Calls++
	type plain Created
	return json.Marshal(plain(v))
}

type Deleted struct {
	ID int `json:"id"`
}

func (*Deleted) event() {}

type Pair struct {
	Left   Event                       `json:"left"`
	Right  Event                       `json:"right"`
	Events []Event                     `json:"events"`
	Extra  jsonschema.Optional[Event]  `json:"extra,omitzero"`
	Note   jsonschema.Optional[string] `json:"note,omitzero"`
}

// Mode deliberately has a misleading String method; names are the wire contract.
type Mode int

const (
	Quiet Mode = -2
	Loud  Mode = 7
)

func (Mode) String() string { return "not-a-wire-name" }

type Config struct {
	Event    Pair                      `json:"event"`
	Direct   Mode                      `json:"direct"`
	Numeric  Mode                      `json:"numeric"`
	Optional jsonschema.Optional[Mode] `json:"optional,omitzero"`
	Nullable jsonschema.Nullable[Mode] `json:"nullable"`
}
