package optional_without_omitzero

import "github.com/tylergannon/polytype"

type Config struct {
	Value polytype.Optional[int] `json:"value"`
}
