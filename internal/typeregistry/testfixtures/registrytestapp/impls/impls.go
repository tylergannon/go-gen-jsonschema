package impls

// priv is a marker interface.
type priv interface {
	SomeMethod()
}

// MyStruct1 represents an entity with a coolness factor.
type MyStruct1 struct {
	// Coolness indicates how cool the entity is on a scale of 1-100.
	Coolness int `json:"coolness"`
}

func (m MyStruct1) SomeMethod() {}

var _ priv = MyStruct1{}

// MyStruct2 represents an entity with a unique name.
type MyStruct2 struct {
	// Name is the unique name of the entity.
	Name string `json:"name"`
}

func (m *MyStruct2) SomeMethod() {}

var _ priv = &MyStruct2{}
