package ptrfixture

//go:generate go run github.com/tylergannon/go-gen-jsonschema/gen-jsonschema

import "encoding/json"

// Thing is registered from a pointer-receiver root: Declare((*Thing).Schema).
type Thing struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func (t *Thing) NameSchema() json.Marshaler {
	return json.RawMessage(`{"type":"string","description":"pointer-root accessor provider ran"}`)
}

func (t *Thing) CountSchema(v int) json.Marshaler {
	return json.RawMessage(`{"type":"integer","description":"pointer-root method provider ran"}`)
}
