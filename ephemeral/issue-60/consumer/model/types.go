package inspectproof

import (
	"encoding/json"
	"net/url"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

type Supported struct {
	Name  string `json:"name"`
	Items []int  `json:"items"`
}

type MyByte uint8

type Unsupported struct {
	Data  []MyByte `json:"data"`
	Count int      `json:"count,string"`
}

type HookValue struct {
	Value string `json:"value"`
}

func (h HookValue) MarshalJSON() ([]byte, error) {
	type plain HookValue
	return json.Marshal(plain(h))
}

type Unknown struct {
	Hook HookValue `json:"hook"`
	URL  url.URL   `json:"url"`
}

type OptionalMissingOmitzero struct {
	Value jsonschema.Optional[int] `json:"value"`
}

type AnyModel struct {
	Payload any `json:"payload"`
}
