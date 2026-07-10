//go:build jsonschema

package remoteenum

import jsonschema "github.com/tylergannon/go-gen-jsonschema"

var _ = jsonschema.NewEnumType[RemoteEnum]()
