//go:build jsonschema
// +build jsonschema

package basictypes

import (
	"encoding/json"
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

func (StringTypeInSharedDecl) Schema() (json.RawMessage, error) {
	panic("not implemented")
}

var (
	_ = jsonschema.NewJSONSchemaMethod(TypeInItsOwnDecl.Schema)
	_ = jsonschema.NewJSONSchemaMethod(TypeInNestedDecl.Schema)
	_ = jsonschema.NewJSONSchemaMethod(TypeInSharedDecl.Schema)
	_ = jsonschema.NewJSONSchemaMethod(StringTypeInSharedDecl.Schema)
)
