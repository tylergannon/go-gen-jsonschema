package defined_wrapper

import "github.com/tylergannon/polytype"

type MaybeInt polytype.Optional[int]

type Config struct {
	Value MaybeInt `json:"value,omitzero"`
}
