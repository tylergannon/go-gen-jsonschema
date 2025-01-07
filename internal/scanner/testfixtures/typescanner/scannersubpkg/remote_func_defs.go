package scannersubpkg

func RemoteFuncOneTypeArg[T any](T) struct{} {
	return struct{}{}
}

func RemoteFuncTwoTypeArg[T any, U any]() struct{} {
	return struct{}{}
}

func RemoteFuncThreeTypeArg[T any, U any, V any](T, U, V) struct{} {
	return struct{}{}
}

type (
	Type001 struct{}
	Type002 struct{}
	Type003 struct{}
	Type004 struct{}
)
