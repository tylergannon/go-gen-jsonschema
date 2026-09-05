//go:build jsonschema

package json

import jsonschema "github.com/tylergannon/go-gen-jsonschema"

var _ = jsonschema.NewInterfaceImpl[Event](Created{})
