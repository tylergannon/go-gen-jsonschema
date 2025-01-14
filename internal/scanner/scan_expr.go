package scanner

// maxTypeDepth helps guard against extremely nested generics, like Abstract[Abstract[...]].
const maxTypeDepth = 5

type TypeKind int

const (
	// NormalConcrete The type named on the TypeID is just the basic or named type.
	NormalConcrete TypeKind = iota
	// Pointer It's a pointer to the basic or named type.
	Pointer
	// SliceOfConcrete means the type is a slice of some named type or basic type.
	SliceOfConcrete
	// SliceOfPointer means the type is a slice of pointer to a named or basic type.
	SliceOfPointer
)

// TypeID is our structured representation of a type. It can represent named types,
// pointers, slices, arrays, and generics—plus marks invalid or disallowed types.
type TypeID struct {
	Kind TypeKind
	// May be empty string if there is no package info available, meaning the type
	// is defined in the current package from where it's referenced.
	PkgPath  string
	TypeName string
	// MUST be set to true if the PkgPath is empty.
	LocalPkg bool
}
