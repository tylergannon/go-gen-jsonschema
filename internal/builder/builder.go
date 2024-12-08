package builder

import (
	"encoding/json"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry"
)

func New(graph *typeregistry.SchemaGraph) *SchemaBuilder {
	return &SchemaBuilder{
		graph:       graph,
		definitions: map[string]json.RawMessage{},
		typeIDMap:   map[typeregistry.TypeID]string{},
	}
}

type SchemaBuilder struct {
	graph *typeregistry.SchemaGraph
	// definitions holds the actual definitions object
	definitions map[string]json.RawMessage
	// typeIDMap maps TypeIDs to the string key in the definitions.
	typeIDMap map[typeregistry.TypeID]string
}

func (b *SchemaBuilder) renderNode(node *typeregistry.Node) json.Marshaler {

	return nil
}
