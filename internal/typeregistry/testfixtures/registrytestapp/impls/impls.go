package impls

// priv is a marker interface.
type priv interface {
	SomeMethod()
}

// NestedStruct1 represents an entity with a coolness factor.
type NestedStruct1 struct {
	// Coolness indicates how cool the entity is on a scale of 1-100.
	Coolness int `json:"coolness"`
}

func (m NestedStruct1) SomeMethod() {}

var _ priv = NestedStruct1{}

// NestedStruct2 represents an entity with a unique name.
type NestedStruct2 struct {
	// Name is the unique name of the entity.
	Name string `json:"name"`
}

func (m *NestedStruct2) SomeMethod() {}

var _ priv = &NestedStruct2{}

// NestedStruct3 represents an entity with a unique name.
type NestedStruct3 struct {
	// Name is the unique name of the entity.
	Name string `json:"name"`
}

func (m *NestedStruct3) SomeMethod() {}

var _ priv = &NestedStruct3{}
