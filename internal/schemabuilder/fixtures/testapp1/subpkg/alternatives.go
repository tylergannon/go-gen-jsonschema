package subpkg

import "time"

type SomeTypeWithTime struct {
	// This annotation will be type checked that:
	// (a) RelativeTime exists.
	// (b) RelativeTime is either:
	//     - Identical to this type (type alias or actually the very type)
	//	   - Defined as a derivative type of this type
	//     - Has a `func ToTime() time.Time` method (not implemented in first version)
	Time          time.Time `json:"time" jsonschema:"LLMFriendlyTime,pkg=github.com/tylergannon/go-gen-jsonschema/types"`
	Length        time.Duration
	SomeInterface SomeInterface `json:"-"`
}

type SomeInterface interface{}
