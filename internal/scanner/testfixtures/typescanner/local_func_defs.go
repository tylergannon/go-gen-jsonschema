package typescanner

func LocalFuncOneTypeArg[T any](T) struct{} {
	return struct{}{}
}

func LocalFuncTwoTypeArg[T any, U any]() struct{} {
	return struct{}{}
}

func LocalFuncThreeTypeArg[T any, U any, V any](T, U, V) struct{} {
	return struct{}{}
}
