// Package compiletest holds small negative compile-time fixtures proving
// that Go itself rejects an unsupported *Declaration[T] chain, per the
// "let the compiler reject it" acceptance bullet for the fluent Declare API
// (issue #73). Fixture packages live under testdata/ (excluded from normal
// `go build ./...`/`go test ./...` by Go's own testdata convention) and are
// each compiled standalone here, asserting only that compilation fails -
// not matching specific compiler error text.
package compiletest

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNegativeFixturesFailToCompile(t *testing.T) {
	t.Parallel()

	cases := []string{
		"mismatched_receiver",       // .Method's provider receiver doesn't match Declare's root receiver
		"mismatched_accessor",       // .Accessor's provider receiver doesn't match Declare's root receiver
		"mismatched_method_field",   // .Method's field type doesn't match its provider's field-value parameter
		"mismatched_function_field", // .Function's field type doesn't match its provider's parameter
	}

	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			out, err := exec.Command("go", "build", "./testdata/"+name+"/").CombinedOutput()
			require.Error(t, err, "expected %s to fail to compile, but it built cleanly:\n%s", name, out)
		})
	}
}
