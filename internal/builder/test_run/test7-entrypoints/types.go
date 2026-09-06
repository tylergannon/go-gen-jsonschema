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

// PointerFuncType is a named pointer type. Go forbids declaring a method on
// it (invalid receiver base type), so its schema entrypoint must be
// registered as a free function.
type PointerFuncType *int
