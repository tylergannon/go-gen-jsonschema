package jsonschema

type UnionType struct{}

type TypeAlt[T any] struct{}

func SetTypeAlternative[T any](typeAlternatives ...TypeAlt[T]) UnionType {
	return UnionType{}
}

func Alt[T any, U any](name string, f func(t T) (U, error)) TypeAlt[U] {
	return struct{}{}
}
