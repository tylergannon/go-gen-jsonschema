package wrapper_in_container

import "github.com/tylergannon/polytype"

type Config struct {
	Values []polytype.Optional[int] `json:"values"`
}
