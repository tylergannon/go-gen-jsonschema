//go:build jsonschema
// +build jsonschema

package enumsremote

import (
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

var _ = jsonschema.NewEnumType[RemoteEnumType]()
