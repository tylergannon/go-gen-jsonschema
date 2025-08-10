package entrypoints

//go:generate go run ./gen

type MethodType struct {
	Name string `json:"name"`
}

type FuncType struct {
	Name string `json:"name"`
}

type BuilderType struct {
	Name string `json:"name"`
}
