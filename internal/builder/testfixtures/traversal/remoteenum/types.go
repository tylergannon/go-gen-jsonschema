package remoteenum

// RemoteEnum is an enum reached directly and through remotestruct.
type RemoteEnum string

const (
	RemoteEnumFirst  RemoteEnum = "first"
	RemoteEnumSecond RemoteEnum = "second"
)

// RemoteEnum declares itself as an enum; the generator emits its typed constants.
func (RemoteEnum) enum() {}

// This package runs no generation of its own, so it references its marked
// enum types by hand; a generated file would emit the same assertions.
var (
	_ interface{ enum() } = *new(RemoteEnum)
)
