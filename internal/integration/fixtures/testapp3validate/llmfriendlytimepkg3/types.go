package llmfriendlytimepkg3

import (
	"time"
)

// LLMFriendlyTime provides union type alternatives for allowing an LLM to choose
// how it specifies the time.
// To account for the LLM's inability to do date arithmetic, it should be able
// to provide actual times in various frames
type LLMFriendlyTime time.Time

type (
	// Choose the unit of time given by the user.
	TimeUnit string

	Month string

	DayOfWeek string

	TimeFrame string
)

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

// TimeFromNow represents a relative time in the future, and is given as
// a time unit and quantity, for instance "7 weeks" (from now).
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

// ActualTime is used when referencing a specific historical date that has
// already occurred or when the full date including year is explicitly known.
// Best for dates in the past or when you need to specify an exact year.
// Examples: 'January 1st, 2004', 'February 14th, 2013', 'July 4th, 1776'
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

const (
	January   Month = "January"   // First month of the year.
	February  Month = "February"  // Second month of the year.
	March     Month = "March"     // Third month of the year.
	April     Month = "April"     // Fourth month of the year.
	May       Month = "May"       // Fifth month of the year.
	June      Month = "June"      // Sixth month of the year.
	July      Month = "July"      // Seventh month of the year.
	August    Month = "August"    // Eighth month of the year.
	September Month = "September" // Ninth month of the year.
	October   Month = "October"   // Tenth month of the year.
	November  Month = "November"  // Eleventh month of the year.
	December  Month = "December"  // Twelfth month of the year.
)
const (
	Sunday    DayOfWeek = "Sunday"
	Monday    DayOfWeek = "Monday"
	Tuesday   DayOfWeek = "Tuesday"
	Wednesday DayOfWeek = "Wednesday"
	Thursday  DayOfWeek = "Thursday"
	Friday    DayOfWeek = "Friday"
	Saturday  DayOfWeek = "Saturday"
)

const (
	Past   TimeFrame = "Past"
	Future TimeFrame = "Future"
)

// NearestDate is used when referencing a calendar date (month and day) without
// a specific year, typically for recurring dates or future dates within the
// current year. Best for dates like 'March 13th', 'December 25th', or
// 'next April 1st'. The timeFrame property specifies whether it's looking into
// the past ('last March 13th') or future ('next March 13th').
type NearestDate struct {
	// TimeFrame specifies whether the reference looks into the future or
	// into the past.
	TimeFrame TimeFrame `json:"timeFrame"`
	Month     Month     `json:"month"`
	// DayOfMonth should be 1 if not specified, otherwise the specific day of the
	// month given by the user.
	DayOfMonth int `json:"dayOfMonth"`
}

// NearestDay is selected when the user has specified a day of the week,
// such as last Friday or next Saturday, or two Wednesdays ago.
type NearestDay struct {
	TimeFrame TimeFrame `json:"timeFrame"`
	DayOfWeek DayOfWeek `json:"dayOfWeek"`
	// Scale is the number of weeks to count. Normally it is one (1).
	// In the example of "two Wednesdays ago" the scale is two (2).
	Scale int `json:"scale"`
}
