package nullable_slice

import "github.com/tylergannon/polytype"

type Config struct {
	Value polytype.Nullable[[]int] `json:"value"`
}
