package builder

import (
	"encoding/json"
	"fmt"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry"
	"log"
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
	rootNode := b.renderNode(b.graph.RootNode)
	if len(b.definitions) == 0 {
		return rootNode
	}
	defs := basicMarshaler(b.definitions)

	switch n := rootNode.(type) {
	case basicMarshaler:
		n[b.definitionsKey] = defs
	case *jsonSchema:
		n.DefinitionsKey = b.definitionsKey
		n.Definitions = defs
		_ = n
		_ = defs
	}
	return rootNode
}

// renderChildNode either returns the schema Marshaler representing the object
// denoted by id, or else an object representing a ref to the definition for the
// object.  If the definition is needed but does not yet exist, it will recurse
// back to renderNode.
func (b *SchemaBuilder) renderChildNode(id typeregistry.TypeID) json.Marshaler {
	if _, found := b.graph.Nodes[id]; !found {
		log.Println("Nodes Count", len(b.graph.Nodes))
		for _id := range b.graph.Nodes {
			fmt.Println(_id)
		}
		panic(fmt.Sprintf("unknown child node %s", id))
	}
	var node = b.graph.Nodes[id]
	switch n := node.(type) {
	case typeregistry.NamedTypeNode, typeregistry.EnumTypeNode:
		if node.Inbound() == 1 {
			return b.renderNode(node)
		}
	case typeregistry.BasicTypeNode:
		return newBasicType(n.BasicType())
	default:
		return b.renderNode(node)
	}
	//if _, ok := node.(typeregistry.NamedTypeNode); !ok || node.Inbound() == 1 {
	//	return b.renderNode(node)
	//} else if basic, ok := node.(typeregistry.BasicTypeNode); ok {
	//	return newBasicType(basic.BasicType())
	//} else if enum, ok := node.(typeregistry.EnumTypeNode); ok {
	//	return newEnumType(enum)
	//}

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

func inspect(str string, item ...any) {
	fmt.Printf("inspect %s: %T %v\n", str, item[0], item[0])
	for i := 1; i < len(item); i++ {
		fmt.Printf("     item[%d]: %T %v\n", i, item[i], item[i])
	}
}

func (b *SchemaBuilder) renderNode(node typeregistry.Node) json.Marshaler {

	switch n := node.(type) {
	case typeregistry.BasicTypeNode:
		return newBasicType(n.BasicType())
	case typeregistry.EnumTypeNode:
		return newEnumType(n)
	case typeregistry.NamedTypeNode:
		child := b.renderChildNode(n.UnderlyingTypeID())
		switch chType := child.(type) {
		case basicMarshaler:
			if rawComments, err := json.Marshal(buildComments(n.TypeSpec.Decorations())); err != nil {
				panic(err)
			} else if len(rawComments) > 0 {
				chType["description"] = json.RawMessage(rawComments)
			}
			return chType
		case *jsonSchema:
			typeComments := buildComments(n.TypeSpec.Decorations())
			if len(chType.Description) == 0 {
				chType.Description = typeComments
			} else {
				chType.Description = typeComments + "\n\n" + chType.Description
			}
			return chType
		case RefElement:
			log.Println("Here we go")
			log.Println(string(chType))
			return chType
		default:
			inspect("Child Type", chType)
		}
	case typeregistry.SliceTypeNode:
		elem := b.renderChildNode(n.ElemNodeID())
		return arraySchema(elem, buildComments(n.DSTNode().Decorations()))
	case typeregistry.StructTypeNode:
		return b.renderStructNode(n)
	default:
		panic(fmt.Sprintf("unknown node type in renderNode %T", n))
	}
	inspect("Render Node", node)

	panic("render node ended unsatisfactorily")
}

func (b *SchemaBuilder) renderStructNode(node typeregistry.StructTypeNode) json.Marshaler {
	schema := &jsonSchema{
		DefinitionsKey: b.definitionsKey,
		Strict:         true,
	}
	var (
		structFieldNodes = make([]typeregistry.StructFieldNode, len(node.Fields()))
		childNodes       = make([]json.Marshaler, len(node.Fields()))
		haveDescription  bool
	)

	for i, field := range node.Fields() {
		structFieldNodes[i] = b.graph.Nodes[field].(typeregistry.StructFieldNode)
		childNodes[i] = b.renderChildNode(structFieldNodes[i].FieldTypeID)
		switch t := childNodes[i].(type) {
		case RefElement:
			haveDescription = true
		case basicMarshaler:
			if _, ok := t["description"]; ok {
				haveDescription = true
			}
		case *jsonSchema:
			if t.Description != "" {
				haveDescription = true
			}
		}
	}
	sb := strings.Builder{}
	sb.WriteString(buildComments(node.DSTNode().Decorations()))
	if haveDescription {
		sb.WriteString("\n\n## **Properties**\n\n")
		for _, field := range structFieldNodes {
			description := buildComments(field.FieldConf.Field.Decorations())
			if len(description) == 0 {
				continue
			}
			description = strings.TrimPrefix(description, field.FieldConf.Var.Name())
			description = strings.TrimSpace(description)
			sb.WriteString("### ")
			sb.WriteString(field.FieldConf.FieldName)
			sb.WriteString("\n\n")
			sb.WriteString(description)
			sb.WriteString("\n\n")
		}
	} else {
		for i, field := range structFieldNodes {
			description := buildComments(field.FieldConf.Decorations())
			if len(description) == 0 {
				continue
			}
			switch t := childNodes[i].(type) {
			case basicMarshaler:
				if rawDescription, err := json.Marshal(description); err != nil {
					panic(err)
				} else if len(rawDescription) > 0 {
					t["description"] = json.RawMessage(rawDescription)
				}
			case *jsonSchema:
				t.Description = description
			}
		}
	}
	for i, field := range structFieldNodes {
		schema.Properties = append(schema.Properties, schemaProperty{
			name: field.FieldConf.FieldName,
			def:  childNodes[i],
		})
	}

	schema.Description = sb.String()

	return schema
}
