package indirectsubpkg

// IntType is Foobarbax
type IntType int

// PointerToIntType is Footballbat
type PointerToIntType *int

// PointerToIntType is a pointer to a named type
type PointerToNamedType *IntType

// DefinedAsNamedType is defined as another named type
type DefinedAsNamedType IntType

// SliceOfPointerToInt is a slice of pointers to int
type SliceOfPointerToInt []*int

// SliceOfPointerToNamedType is a slice of pointers to named types
type SliceOfPointerToNamedType []*IntType

// SliceOfNamedType is a slice of a named type
type SliceOfNamedType []IntType

// NamedSliceType is a slice of a named type.
type NamedSliceType SliceOfNamedType

// NamedNamedSliceType is another level of indirection from NamedSliceType
type NamedNamedSliceType NamedSliceType

// SliceOfNamedNamedSliceType is a slice of the named slice type.
type SliceOfNamedNamedSliceType []NamedNamedSliceType
