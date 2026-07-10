package remoteenum

// RemoteEnum is an enum reached directly and through remotestruct.
type RemoteEnum string

const (
	RemoteEnumFirst  RemoteEnum = "first"
	RemoteEnumSecond RemoteEnum = "second"
)
