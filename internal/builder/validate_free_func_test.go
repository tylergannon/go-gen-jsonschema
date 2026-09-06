package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestValidateRejectsFreeFunctionPointerRoot proves that --validate fails
// fast, with an actionable error, instead of silently succeeding while
// omitting ValidateJSON for a free-function schema root whose type can't
// have a method declared on it (pointer/interface underlying type). Before
// this check, generation would succeed and simply produce no ValidateJSON
// for that type -- a silent partial result.
func TestValidateRejectsFreeFunctionPointerRoot(t *testing.T) {
	dir := writeTypeGrammarFixture(t, `//go:build jsonschema

package fixture

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

type PointerRoot *int

func PointerRootSchema(PointerRoot) json.RawMessage { panic("not implemented") }

var _ = polytype.Declare(PointerRootSchema)
`)

	err := Run(BuilderArgs{TargetDir: dir, Validate: true})
	require.ErrorContains(t, err, "--validate cannot generate ValidateJSON for PointerRoot")
}
