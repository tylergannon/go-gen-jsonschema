package nullable_interface

import "github.com/tylergannon/polytype"

type Value interface{ value() }

type Text struct {
	Text string `json:"text"`
}

func (Text) value() {}

type Config struct {
	Value polytype.Nullable[Value] `json:"value"`
}
