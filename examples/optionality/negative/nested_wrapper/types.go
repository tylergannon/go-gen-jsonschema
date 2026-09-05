package nested_wrapper

import "github.com/tylergannon/polytype"

type Config struct {
	Value polytype.Optional[polytype.Optional[int]] `json:"value,omitzero"`
}
