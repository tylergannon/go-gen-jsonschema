package scaffold_demo

import (
	"encoding/json"
	"fmt"
)

func Demo() string {
	w := Widget{}
	var v map[string]any
	if err := json.Unmarshal(w.Schema(), &v); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%v", v["required"])
}
