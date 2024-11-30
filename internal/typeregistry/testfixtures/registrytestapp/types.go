package registrytestapp

import (
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry/testfixtures/registrytestapp/subpkg"
)

var _ = jsonschema.SetTypeAlternative[subpkg.LLMFriendlyTime](
	// For referencing a time in the past using relative units
	jsonschema.Alt("timeAgo", subpkg.TimeAgo.ToTime),
	// For referencing a time in the future using relative units
	jsonschema.Alt("timeFromNow", subpkg.TimeFromNow.ToTime),
	// When given an actual time. Must be valid RFC3339 time format
	jsonschema.Alt("actualTime", subpkg.ActualTime.ToTime),
	// To refer to the present moment
	jsonschema.Alt("now", subpkg.Now.ToTime),
	// To reference all of history
	jsonschema.Alt("beginningOfTime", subpkg.BeginningOfTime.ToTime),
)
