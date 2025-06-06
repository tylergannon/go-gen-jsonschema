package basictypes

//go:generate go run ./gen

import (
	"github.com/tylergannon/go-gen-jsonschema/internal/builder/testfixtures/enums/enumsremote"
)

// EnumType is an enum type from enumsremote
type EnumType string

const (
	// EnumVal1 is a value!!
	EnumVal1 EnumType = "val1"
	// EnumVal2 is also a value!!
	EnumVal2 EnumType = "val2"
	// EnumVal3 is truly a value!!
	EnumVal3 EnumType = "val3"
)

// EnumVal4 is the fourth value
const EnumVal4 EnumType = "val4"

// SliceOfEnumType is a slice of the enums.
type SliceOfEnumType []EnumType

// SliceOfRemoteEnumType is a slice of the remote enum type
type SliceOfRemoteEnumType []enumsremote.RemoteEnumType

// SliceOfPointerToRemoteEnum is a slice of pointers to the remote enum type
type SliceOfPointerToRemoteEnum []*enumsremote.RemoteEnumType
