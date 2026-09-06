package asref_collision_dep_1423005799

import "encoding/json"

type Shared struct {
	Name string `json:"name"`
}

func (Shared) Schema() json.RawMessage { panic("not implemented") }
