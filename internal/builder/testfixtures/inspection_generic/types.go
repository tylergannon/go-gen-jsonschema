package inspection_generic

type Box[T any] struct {
	Value T `json:"value"`
}

type Root struct {
	Box[string]
}
