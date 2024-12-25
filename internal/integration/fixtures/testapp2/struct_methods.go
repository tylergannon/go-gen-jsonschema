package testapp2

import (
	"errors"
	"fmt"
	"github.com/tylergannon/go-gen-jsonschema-testapp/llmfriendlytimepkg"
	"time"
)

// parseMonth converts our Month enum to a time.Month (1–12).
func parseMonth(m llmfriendlytimepkg.Month) (time.Month, error) {
	switch m {
	case llmfriendlytimepkg.January:
		return time.January, nil
	case llmfriendlytimepkg.February:
		return time.February, nil
	case llmfriendlytimepkg.March:
		return time.March, nil
	case llmfriendlytimepkg.April:
		return time.April, nil
	case llmfriendlytimepkg.May:
		return time.May, nil
	case llmfriendlytimepkg.June:
		return time.June, nil
	case llmfriendlytimepkg.July:
		return time.July, nil
	case llmfriendlytimepkg.August:
		return time.August, nil
	case llmfriendlytimepkg.September:
		return time.September, nil
	case llmfriendlytimepkg.October:
		return time.October, nil
	case llmfriendlytimepkg.November:
		return time.November, nil
	case llmfriendlytimepkg.December:
		return time.December, nil
	}
	return time.January, fmt.Errorf("unknown month %q", m)
}

// parseDayOfWeek converts our DayOfWeek enum into time.Weekday (0=Sunday).
func parseDayOfWeek(d llmfriendlytimepkg.DayOfWeek) (time.Weekday, error) {
	switch d {
	case llmfriendlytimepkg.Sunday:
		return time.Sunday, nil
	case llmfriendlytimepkg.Monday:
		return time.Monday, nil
	case llmfriendlytimepkg.Tuesday:
		return time.Tuesday, nil
	case llmfriendlytimepkg.Wednesday:
		return time.Wednesday, nil
	case llmfriendlytimepkg.Thursday:
		return time.Thursday, nil
	case llmfriendlytimepkg.Friday:
		return time.Friday, nil
	case llmfriendlytimepkg.Saturday:
		return time.Saturday, nil
	}
	return time.Sunday, fmt.Errorf("unknown day of week %q", d)
}

// ToTime calculates the nearest date (month + day) in the future or past.
func NearestDateToTime(nd llmfriendlytimepkg.NearestDate) (LLMFriendlyTime, error) {
	now := time.Now()

	// Resolve the month to time.Month
	tmMonth, err := parseMonth(nd.Month)
	if err != nil {
		return LLMFriendlyTime{}, err
	}

	// Default DayOfMonth to 1 if unset or invalid
	dayOfMonth := nd.DayOfMonth
	if dayOfMonth <= 0 || dayOfMonth > 31 {
		dayOfMonth = 1
	}

	// Build a candidate time for the current year
	loc := now.Location()
	candidate := time.Date(now.Year(), tmMonth, dayOfMonth, 0, 0, 0, 0, loc)

	switch nd.TimeFrame {
	case llmfriendlytimepkg.Future:
		// If candidate is before or equal to 'now', move it to next year's occurrence
		if !candidate.After(now) {
			candidate = candidate.AddDate(1, 0, 0)
		}
	case llmfriendlytimepkg.Past:
		// If candidate is after or equal to 'now', move it to last year's occurrence
		if !candidate.Before(now) {
			candidate = candidate.AddDate(-1, 0, 0)
		}
	default:
		return LLMFriendlyTime{}, errors.New("invalid TimeFrame: must be Past or Future")
	}

	return LLMFriendlyTime(candidate), nil
}

// ToTime calculates the nearest day-of-week in the future or past.
// The Scale indicates how many weeks to move.
func NearestDayToTime(nd llmfriendlytimepkg.NearestDay) (LLMFriendlyTime, error) {
	now := time.Now()

	w, err := parseDayOfWeek(nd.DayOfWeek)
	if err != nil {
		return LLMFriendlyTime{}, err
	}

	// Default Scale to 1 if not provided or invalid
	scale := nd.Scale
	if scale < 1 {
		scale = 1
	}

	// Helper function: nextDayOfWeek returns the next occurrence of w strictly after 'start'.
	nextDayOfWeek := func(start time.Time, w time.Weekday) time.Time {
		// Convert time.Weekday (Sunday=0) to int
		current := int(start.Weekday())
		target := int(w)

		// Days until next occurrence
		daysUntil := (target - current + 7) % 7
		if daysUntil == 0 {
			// If it's already that day, we move 7 days ahead for “next”
			daysUntil = 7
		}

		return start.AddDate(0, 0, daysUntil)
	}

	// Helper function: prevDayOfWeek returns the last occurrence of w strictly before 'start'.
	prevDayOfWeek := func(start time.Time, w time.Weekday) time.Time {
		// Convert time.Weekday (Sunday=0) to int
		current := int(start.Weekday())
		target := int(w)

		// Days until last occurrence
		daysUntil := (current - target + 7) % 7
		if daysUntil == 0 {
			// If it's already that day, we go 7 days back for “previous”
			daysUntil = 7
		}

		return start.AddDate(0, 0, -daysUntil)
	}

	var candidate time.Time
	switch nd.TimeFrame {
	case llmfriendlytimepkg.Future:
		// Find the next occurrence
		candidate = nextDayOfWeek(now, w)
		// If Scale > 1, then skip that many additional weeks
		if scale > 1 {
			candidate = candidate.AddDate(0, 0, 7*(scale-1))
		}
	case llmfriendlytimepkg.Past:
		// Find the last occurrence
		candidate = prevDayOfWeek(now, w)
		// If Scale > 1, go that many additional weeks back
		if scale > 1 {
			candidate = candidate.AddDate(0, 0, -7*(scale-1))
		}
	default:
		return LLMFriendlyTime{}, errors.New("invalid TimeFrame: must be Past or Future")
	}

	return LLMFriendlyTime(candidate), nil
}
