package remoteenum

// RemoteEnum is an enum reached directly and through remotestruct.
type RemoteEnum string

const (
	RemoteEnumFirst  RemoteEnum = "first"
	RemoteEnumSecond RemoteEnum = "second"
)

// RemoteEnum declares itself as an enum; the generator emits its typed constants.
//
//lint:ignore U1000 enum marker method, read by the polytype generator
func (RemoteEnum) enum() {}
