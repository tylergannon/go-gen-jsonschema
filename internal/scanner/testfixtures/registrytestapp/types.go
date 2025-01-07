package registrytestapp

import (
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry/testfixtures/registrytestapp/subpkg"
	"time"
)

type (
	LLMFriendlyTime subpkg.LLMFriendlyTime
)

func TimeAgoToLLMFriendlyTime(t subpkg.TimeAgo) (LLMFriendlyTime, error) {
	return LLMFriendlyTime(time.Now().Add(-subpkg.ToDuration(t.Unit, t.Quantity))), nil
}

func FromNowToLLMFriendlyTime(t subpkg.TimeFromNow) (LLMFriendlyTime, error) {
	return LLMFriendlyTime(time.Now().Add(subpkg.ToDuration(t.Unit, t.Value))), nil
}

func ActualTimeToLLMFriendlyTime(t subpkg.ActualTime) (LLMFriendlyTime, error) {
	_t, err := t.ToTime()
	return LLMFriendlyTime(_t), err
}

func NowToLLMFriendlyTime(t subpkg.Now) (LLMFriendlyTime, error) {
	_t, err := t.ToTime()
	return LLMFriendlyTime(_t), err
}
func BeginningOfTimeToLLMFriendlyTime(t subpkg.BeginningOfTime) (LLMFriendlyTime, error) {
	_t, err := t.ToTime()
	return LLMFriendlyTime(_t), err
}

var _ = jsonschema.SetTypeAlternative[LLMFriendlyTime](
	// For referencing a time in the past using relative units
	jsonschema.Alt("timeAgo", TimeAgoToLLMFriendlyTime),
	// For referencing a time in the future using relative units
	jsonschema.Alt("timeFromNow", FromNowToLLMFriendlyTime),
	// When given an actual time. Must be valid RFC3339 time format
	jsonschema.Alt("actualTime", ActualTimeToLLMFriendlyTime),
	// To refer to the present moment
	jsonschema.Alt("now", NowToLLMFriendlyTime),
	// To reference all of history
	jsonschema.Alt("beginningOfTime", BeginningOfTimeToLLMFriendlyTime),
)
