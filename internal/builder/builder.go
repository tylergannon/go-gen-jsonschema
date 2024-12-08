package builder

import (
	"encoding/json"
	"fmt"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry"
	"go/types"
	"strconv"
	"strings"
)

func New(graph *typeregistry.SchemaGraph) *SchemaBuilder {
	return &SchemaBuilder{
		graph:          graph,
		definitions:    definitionsMap{},
		typeIDMap:      map[typeregistry.TypeID]string{},
		definitionsKey: defaultDefinitionsKey,
	}
}

type definitionsMap map[string]json.Marshaler

func (d definitionsMap) findNewName(name string) string {
	if _, ok := d[name]; !ok {
		return name
	}
	for i := 0; ; i++ {
		if _, ok := d[name+strconv.Itoa(i)]; !ok {
			return name
		}
	}
}

type SchemaBuilder struct {
	graph *typeregistry.SchemaGraph
	// definitions holds the actual definitions object
	definitions definitionsMap
	// typeIDMap maps TypeIDs to the string key in the definitions.
	typeIDMap      map[typeregistry.TypeID]string
	definitionsKey string
}

func (b *SchemaBuilder) Render() json.Marshaler {
	return b.renderNode(b.graph.RootNode)
}

// renderChildNode either returns the schema Marshaler representing the object
// denoted by id, or else an object representing a ref to the definition for the
// object.  If the definition is needed but does not yet exist, it will recurse
// back to renderNode.
func (b *SchemaBuilder) renderChildNode(id typeregistry.TypeID) json.Marshaler {
	if _, found := b.graph.Nodes[id]; !found {
		panic(fmt.Sprintf("unknown child node %s", id))
	}
	var node = b.graph.Nodes[id]
	if len(node.Children) <= 1 || node.Inbound <= 1 {
		return b.renderNode(node)
	}
	if _, found := b.typeIDMap[id]; !found {
		// prevent infinite recursion by ensuring that the typeIDMap contains
		// this node's ID.  Therefore a cyclic reference will terminate when
		// it reaches here.
		parts := strings.Split(string(id), ".")
		name := b.definitions.findNewName(parts[len(parts)-1])
		b.typeIDMap[id] = name
		b.definitions[name] = b.renderNode(node)
	}
	var ref = b.typeIDMap[id]
	return refElement(fmt.Sprintf("#/%s/%s", b.definitionsKey, ref))
}

func (b *SchemaBuilder) renderNode(node *typeregistry.Node) json.Marshaler {

	// Okay there are only a few options for what the node might have in it.
	switch t := node.Type.(type) {
	case *types.Pointer, *types.Interface, *types.Chan, *types.Signature:
		panic("this really should be impossible")
	case *types.Map:
		panic("map types not yet supported, not sure why")
	case *types.Basic:
		return newBasicType(t)
	}
	var children = map[typeregistry.TypeID]json.Marshaler{}
	for _, child := range node.Children {
		children[child] = b.renderChildNode(child)
	}
	switch t := node.Type.(type) {
	case *types.Named:
		if ch, ok := children[node.Children[0]]; !ok {
			panic(fmt.Sprintf("child node %s is not a child node", node.Children[0]))
		} else {
			switch chType := ch.(type) {
			case basicMarshaler:
				chType["description"] = buildComments(node.Node.Decorations())
				return chType
			}
		}
	case *types.Slice, *types.Array:
	case *types.Struct:
		_ = t
	default:
		panic(fmt.Sprintf("unknown node type in renderNode %T", t))
	}
	panic("render node ended unsatisfactorily")
}

//func (b *SchemaBuilder) renderStructNode(node *typeregistry.Node) json.Marshaler {
//	var (
//		sb         = strings.Builder{}
//		structType = node.Node.(*dst.StructType)
//	)
//
//	return nil
//}
