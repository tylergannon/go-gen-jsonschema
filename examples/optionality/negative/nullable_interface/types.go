package nullable_interface

import jsonschema "github.com/tylergannon/go-gen-jsonschema"

type Value interface{ value() }

type Text struct {
	Text string `json:"text"`
}

func (Text) value() {}

type Config struct {
	Value jsonschema.Nullable[Value] `json:"value"`
}
