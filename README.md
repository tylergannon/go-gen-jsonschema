# go-gen-jsonschema

- Builds simple JSON schemas from static analysis of structs.
- Implements the subset of JSON Schema supported by OpenAI structured outputs.
   This is documented in `Requirements.txt`.
- Obeys the field names given in the `json:"someField"` annotation.
- Reads comments from code immediately before structs, and uses that as the
  description of any JSON Schema object (root or nested)
- Reads the comments from code immediately before field definitions, and uses
  them as the description for properties on object definitions for the schema.
- Follows fields across package boundaries, to their type declaration.
- Handles pointer fields, derivative types, and type aliases.

## Does *not* support

If a struct contains any of the following fields, they *MUST* be marked
`json:"-"` (aka *ignore*) or else generation will fail and no schema will be
emitted:

1. Private (unpublished) fields
2. Interface objects
3. Function object
4. Channel
5. `sync.Mutex`, `sync.Cond`, `sync.WaitGroup` etc

## Usage:

Types indicated on command line must be present in local package.

```go
package mypackage

//go:generate go-gen-jsonschema -type MyType,MyOtherType,MyThirdType
```

## Type Alternatives

We have a somewhat "hot" take on how to enable union types.

To enable it, the type that you'll want to represent as a Union type must be a
named type.  E.g.

### Declare a type

```go
package my_test
import "time"

// LLMFriendlyTime represents a moment in time based on user input.
// Takes any of several forms, depending on whether the user queries about
// (1) an actual date, e.g. "July 1" or "2020-01-01 0800 PST"
// (2) a relative time, e.g. "two days ago", "3 years from now"
// (3) an anchor, e.g. "now" or "beginning of time"
type LLMFriendlyTime time.Time
```

### Create structs that represent the possible types

For each possible representation, provide a struct type that will receive the
representation, and a function that converts from that struct type to your
union type.  Let's fill out our example a little further:

```go
package my_test

import (
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"time"
)

// LLMFriendlyTime represents a moment in time based on user input.
// Takes any of several forms, depending on whether the user queries about
// (1) an actual date, e.g. "July 1" or "2020-01-01 0800 PST"
// (2) a relative time, e.g. "two days ago", "3 years from now"
// (3) an anchor, e.g. "now" or "beginning of time"
type LLMFriendlyTime time.Time

// Choose the unit of time given by the user.
type TimeUnit string

const (
	Minutes TimeUnit = "minutes"
	Hours   TimeUnit = "hours"
	Weeks   TimeUnit = "weeks"
	Days    TimeUnit = "days"
	Months  TimeUnit = "months"
	Years   TimeUnit = "years"
)

// ActualTime is used when the user provides an actual date/date+time.
type ActualTime struct {
	// DateTime must be provided in RFC3339 format.
	// Use UTC if no time zone is given.
	DateTime string `json:"dateTime"`
}

func (t ActualTime) ToTime() (LLMFriendlyTime, error) {
	var theTime time.Time
	err := theTime.UnmarshalText([]byte(t.DateTime))
	return LLMFriendlyTime(theTime), err
}

// TimeAgo represents a relative time in the past.  No need for date
// arithmetic.  Instead, simply specify a time unit and quantity.
// An alternate version of this might actually accept an array of these
// objects, in order to support complex times like "three years, three days,
// three minutes and three seconds ago".
type TimeAgo struct {
	// Choose the unit of as given.
	Unit TimeUnit `json:"unit"`
	// Enter the number of the selected unit.
	Quantity int `json:"quantity"`
}

func (t TimeAgo) ToTime() (LLMFriendlyTime, error) {
	return LLMFriendlyTime(time.Now().Add(-ToDuration(t.Unit, t.Quantity))), nil
}

func ToDuration(unit TimeUnit, value int) time.Duration {
	var dur time.Duration
	switch unit {
	case Minutes:
		dur = time.Minute * time.Duration(value)
	case Hours:
		dur = time.Hour * time.Duration(value)
	case Weeks:
		dur = time.Hour * 24 * 7 * time.Duration(value)
	case Days:
		dur = time.Hour * 24 * time.Duration(value)
	case Months:
		dur = time.Hour * 24 * 30 * time.Duration(value)
	case Years:
		dur = time.Hour * 24 * 365 * time.Duration(value)
	}
	return dur
}


```

### Declare the union types

In this example we use struct methods to provide the conversion, but you can
also use regular non-struct methods, as long as the function receives
a value of your type-alt struct, and outputs a tuple of the target type and
error.

```go
package my_test

import jsonschema "github.com/tylergannon/go-gen-jsonschema"

var _ = jsonschema.SetTypeAlternative[LLMFriendlyTime](
	TimeAgo.ToTime,
	ActualTime.ToTime,
)
```
