package inspection_nested

import "encoding/json"

type Event interface {
	isEvent()
}

type Created struct {
	Name string `json:"name"`
}

func (Created) isEvent() {}

type State int

const (
	StateOpen State = iota + 1
	StateClosed
)

func (s State) String() string {
	if s == StateOpen {
		return "open"
	}
	return "closed"
}

type Child struct {
	Event Event `json:"event"`
	State State `json:"state"`
}

type Parent struct {
	Child Child `json:"child"`
}

type HookValue struct {
	Value string `json:"value"`
}

func (h HookValue) MarshalJSON() ([]byte, error) {
	type plain HookValue
	return json.Marshal(plain(h))
}

type HookModel struct {
	Hook HookValue `json:"hook"`
}

type WireMismatch struct {
	Bytes   []byte   `json:"bytes"`
	Aliased []MyByte `json:"aliased"`
	Count   int      `json:"count,string"`
	Inline  struct {
		Ignored []byte `json:"-"`
		Count   int    `json:"count,string"`
	} `json:"inline"`
}

type MyByte uint8

type ProviderModel struct {
	Value string `json:"value"`
}

func ProviderSchema(string) json.Marshaler {
	return json.RawMessage(`{"type":"string"}`)
}
