package animals

// Animal is sealed by isAnimal; its discriminator may only be declared in
// this package.
type Animal interface {
	isAnimal()
}

type Dog struct {
	Name string `json:"name"`
}

func (Dog) isAnimal() {}
