package enumsremote

// RemoteEnumType is an enum type from enumsremote
type RemoteEnumType string

const (
	// EnumVal1 is a value!!
	EnumVal1 RemoteEnumType = "val1"
	// EnumVal2 is also a value!!
	EnumVal2 RemoteEnumType = "val2"
	// EnumVal3 is truly a value!!
	EnumVal3 RemoteEnumType = "val3"
)

// EnumVal4 is the fourth value
const EnumVal4 RemoteEnumType = "val4"

// RemoteEnumType declares itself as an enum; the generator emits its typed constants.
func (RemoteEnumType) enum() {}

// This package runs no generation of its own, so it references its marked
// enum types by hand; a generated file would emit the same assertions.
var (
	_ interface{ enum() } = *new(RemoteEnumType)
)
