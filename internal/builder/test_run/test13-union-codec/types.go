package union_codec

import (
	"encoding/json"
	"errors"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

//go:generate go run ./gen

type Event interface{ isEvent() }

type Created struct {
	Name string `json:"name"`
}

func (Created) isEvent() {}

type Deleted struct {
	ID string `json:"id"`
}

func (*Deleted) isEvent() {}

type Unregistered struct {
	Value string `json:"value"`
}

func (Unregistered) isEvent() {}

var hookMarshalCalls int
var ordinaryMarshalCalls int

type Hooked struct {
	Name             string `json:"name"`
	Behavior         string `json:"-"`
	SawDiscriminator bool   `json:"-"`
}

func (Hooked) isEvent() {}

func (h Hooked) MarshalJSON() ([]byte, error) {
	hookMarshalCalls++
	switch h.Behavior {
	case "matching":
		return json.Marshal(map[string]any{"hookKind": "hooked", "name": h.Name})
	case "conflict":
		return json.Marshal(map[string]any{"hookKind": "other", "name": h.Name})
	case "non-string":
		return json.Marshal(map[string]any{"hookKind": 3, "name": h.Name})
	case "null":
		return []byte("null"), nil
	case "array":
		return []byte("[]"), nil
	case "string":
		return []byte(`"payload"`), nil
	case "malformed":
		return []byte("{"), nil
	case "error":
		return nil, errors.New("hook failed")
	default:
		return json.Marshal(struct {
			Name string `json:"name"`
		}{Name: h.Name})
	}
}

func (h *Hooked) UnmarshalJSON(data []byte) error {
	var wire struct {
		Kind string `json:"hookKind"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	if wire.Kind != "hooked" {
		return errors.New("Hooked.UnmarshalJSON did not receive the registered discriminator")
	}
	*h = Hooked{Name: wire.Name, SawDiscriminator: true}
	return nil
}

type Ordinary struct {
	Value string `json:"value"`
}

func (o Ordinary) MarshalJSON() ([]byte, error) {
	ordinaryMarshalCalls++
	type alias Ordinary
	return json.Marshal(alias(o))
}

type Nested struct {
	Event Event `json:"event"`
}

type Envelope struct {
	Primary   Event                       `json:"primary"`
	Events    []Event                     `json:"events"`
	Optional  jsonschema.Optional[Event]  `json:"optional,omitzero"`
	Alternate jsonschema.Optional[Event]  `json:"alternate,omitzero"`
	Single    jsonschema.Optional[Event]  `json:"single,omitzero"`
	Hook      jsonschema.Optional[Event]  `json:"hook,omitzero"`
	Nested    Nested                      `json:"nested"`
	Ordinary  Ordinary                    `json:"ordinary"`
	Label     string                      `json:"label"`
	Omitted   jsonschema.Optional[string] `json:"omitted,omitzero"`
}
