//go:build jsonschema

package remoteenum

import "github.com/tylergannon/polytype"

var _ = polytype.NewEnumType[RemoteEnum]()
