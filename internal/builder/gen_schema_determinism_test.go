package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSortedCustomTypeNames(t *testing.T) {
	customTypes := map[string][]InterfaceProp{
		"Payment": nil,
		"Drawing": nil,
		"Account": nil,
	}

	require.Equal(t, []string{"Account", "Drawing", "Payment"}, sortedCustomTypeNames(customTypes))
}
