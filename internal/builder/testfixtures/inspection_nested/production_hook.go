//go:build !jsonschema

package inspection_nested

import "encoding/json"

func (h HookValue) MarshalJSON() ([]byte, error) {
	type plain HookValue
	return json.Marshal(plain(h))
}
