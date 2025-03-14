package syntax

import "fmt"

type Indirection int

const (
	// NormalConcrete The type named on the TypeID is just the basic or named type.
	NormalConcrete Indirection = iota
	// Pointer It's a pointer to the basic or named type.
	Pointer
	// SliceOfConcrete means the type is a slice of some named type or basic type.
	// NOTE: hypothetical.  not currently supported.
	SliceOfConcrete
	// SliceOfPointer means the type is a slice of pointer to a named or basic type.
	// NOTE: hypothetical.  not currently supported.
	SliceOfPointer
)

type TypeID struct {
	// May be empty string if there is no package info available, meaning the type
	// is defined in the current package from where it's referenced.
	PkgPath     string
	TypeName    string
	Indirection Indirection
}

func (t TypeID) Concrete() TypeID {
	var newTypeID = t
	t.Indirection = NormalConcrete
	return newTypeID
}

func (t TypeID) String() string {
	var (
		ptr     string
		pkgPath = t.PkgPath
	)
	if t.Indirection == Pointer {
		ptr = "*"
	}
	if pkgPath == "" {
		panic("pkgPath is empty")
	}
	return fmt.Sprintf("%s%s.%s", ptr, pkgPath, t.TypeName)
}
