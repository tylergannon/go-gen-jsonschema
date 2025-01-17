package subpkg

type Foo int

type Baz struct {
	A Foo
	B Foo
	C Bar
}

type Bar struct {
	A Foo
}
