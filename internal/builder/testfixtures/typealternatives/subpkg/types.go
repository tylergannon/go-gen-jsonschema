package subpkg

import (
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"time"
)

func Norgle(x int) (LLMFriendlyTime, error) {
	return LLMFriendlyTime(time.Now()), nil
}

// LLMFriendlyTime provides union type alternatives for allowing an LLM to choose
// how it specifies the time.
// To account for the LLM's inability to do date arithmetic, it should be able
// to provide actual times in various frames
type LLMFriendlyTime time.Time

var _ = jsonschema.SetTypeAlternative[LLMFriendlyTime](
	// For referencing a time in the past using relative units
	jsonschema.Alt("timeAgo", TimeAgo.ToTime),
	// For referencing a time in the future using relative units
	jsonschema.Alt("timeFromNow", TimeFromNow.ToTime),
	// When given an actual time. Must be valid RFC3339 time format
	jsonschema.Alt("actualTime", ActualTime.ToTime),
	// To refer to the present moment
	jsonschema.Alt("now", Now.ToTime),
	// To reference all of history
	jsonschema.Alt("beginningOfTime", BeginningOfTime.ToTime),
	jsonschema.Alt("norgle", Norgle),
)

var NamedUnionType = jsonschema.SetTypeAlternative[LLMFriendlyTime](
	// For referencing a time in the past using relative units
	jsonschema.Alt("timeAgo", TimeAgo.ToTime),
	// For referencing a time in the future using relative units
	jsonschema.Alt("timeFromNow", TimeFromNow.ToTime),
	// When given an actual time. Must be valid RFC3339 time format
	jsonschema.Alt("actualTime", ActualTime.ToTime),
	// To refer to the present moment
	jsonschema.Alt("now", Now.ToTime),
	// To reference all of history
	jsonschema.Alt("beginningOfTime", BeginningOfTime.ToTime),
)

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

// TimeAgo reflects a relative time in the past, given in units of time
// relative to the present time.
type TimeAgo struct {
	// Choose the unit of as given.
	Unit TimeUnit `json:"unit"`
	// Enter the number of the selected unit.
	Quantity int `json:"quantity"`
}

func (t TimeAgo) ToTime() (LLMFriendlyTime, error) {
	return LLMFriendlyTime(time.Now().Add(-ToDuration(t.Unit, t.Quantity))), nil
}

type TimeFromNow struct {
	// Choose the unit of as given.
	Unit TimeUnit `json:"unit"`
	// Enter the number of the selected unit.
	Value int `json:"value"`
}

func (t TimeFromNow) ToTime() (LLMFriendlyTime, error) {
	return LLMFriendlyTime(time.Now().Add(ToDuration(t.Unit, t.Value))), nil
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

type ActualTime struct {
	DateTime string `json:"dateTime"`
}

func (t ActualTime) ToTime() (LLMFriendlyTime, error) {
	var theTime time.Time
	err := theTime.UnmarshalText([]byte(t.DateTime))
	return LLMFriendlyTime(theTime), err
}

type Now struct{}

func (t Now) ToTime() (LLMFriendlyTime, error) {
	return LLMFriendlyTime(time.Now()), nil
}

type BeginningOfTime struct{}

func (t BeginningOfTime) ToTime() (LLMFriendlyTime, error) {
	return LLMFriendlyTime(time.Unix(0, 0)), nil
}
