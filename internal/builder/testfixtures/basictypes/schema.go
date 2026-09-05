//go:build jsonschema

package basictypes

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

func (TypeInItsOwnDecl) Schema() json.RawMessage {
	panic("not implemented")
}

func (TypeInNestedDecl) Schema() json.RawMessage {
	panic("not implemented")
}

func (TypeInSharedDecl) Schema() json.RawMessage {
	panic("not implemented")
}

func (StringTypeInSharedDecl) Schema() json.RawMessage {
	panic("not implemented")
}

var (
	_ = polytype.NewJSONSchemaMethod(TypeInItsOwnDecl.Schema)
	_ = polytype.NewJSONSchemaMethod(TypeInNestedDecl.Schema)
	_ = polytype.NewJSONSchemaMethod(TypeInSharedDecl.Schema)
	_ = polytype.NewJSONSchemaMethod(StringTypeInSharedDecl.Schema)
)
