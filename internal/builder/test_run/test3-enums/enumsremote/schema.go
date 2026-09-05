//go:build jsonschema

package enumsremote

import (
	"github.com/tylergannon/polytype"
)

var _ = polytype.NewEnumType[RemoteEnumType]()
