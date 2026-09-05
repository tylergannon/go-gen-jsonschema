package nullable_ref

import "github.com/tylergannon/polytype"

type Config struct {
	Value polytype.Nullable[string] `json:"value" jsonschema:"ref=definitions/Value"`
}
