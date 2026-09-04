package inspectproof

import (
	"encoding/json"
	"net/url"
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
