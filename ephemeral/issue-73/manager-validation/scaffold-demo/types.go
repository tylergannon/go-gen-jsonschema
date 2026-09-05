package scaffold_demo

//go:generate go run github.com/tylergannon/go-gen-jsonschema/gen-jsonschema

type Widget struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}
