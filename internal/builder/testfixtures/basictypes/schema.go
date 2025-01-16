//go:build jsonschema
// +build jsonschema

package basictypes

import (
	"encoding/json"
	_ "github.com/dave/dst"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

func (TypeInItsOwnDecl) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (TypeInNestedDecl) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

func (TypeInSharedDecl) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

var (
	_ = jsonschema.NewJSONSchemaMethod(TypeInItsOwnDecl.Schema)
	_ = jsonschema.NewJSONSchemaMethod(TypeInNestedDecl.Schema)
	_ = jsonschema.NewJSONSchemaMethod(TypeInSharedDecl.Schema)
)
