// nolint
package testapp0_simple

type InvalidDueToFunctionField1 struct {
	Func func()
}

type InvalidDueToFunctionField2 struct {
	Func *func()
}

type InvalidDueToChannelField1 struct {
	Chan chan int
}

type InvalidDueToChannelField2 struct {
	Chan *chan int
}

type Foobar interface {
	Nice() int
}

type InvalidDueToInterfaceField1 struct {
	Field1 any
}

type InvalidDueToInterfaceField2 struct {
	Field2 Foobar
}

type InvalidDueToInterfaceField3 struct {
	Field2 *Foobar
}

type InvalidDueToPrivateField struct {
	privateField int
}
