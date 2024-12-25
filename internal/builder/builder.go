package builder

import (
	"encoding/json"
	"fmt"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry"
	"log"
	"strconv"
	"strings"
)

const (
	discriminatorPropName = "__type__"
)

func New(graph *typeregistry.SchemaGraph) *SchemaBuilder {
	return &SchemaBuilder{
		graph:                 graph,
		definitions:           definitionsMap{},
		typeIDMap:             map[typeregistry.TypeID]string{},
		definitionsKey:        defaultDefinitionsKey,
		DiscriminatorPropName: discriminatorPropName,
		discriminators:        map[string]bool{},
		discriminated:         map[typeregistry.TypeID]string{},
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
	typeIDMap             map[typeregistry.TypeID]string
	definitionsKey        string
	DiscriminatorPropName string
	discriminators        map[string]bool
	discriminated         map[typeregistry.TypeID]string
}

/*
For rendering the json marshalers for types with alternatives.
We only do that within the same package, right?
No?
I think that for now we need to make one file per type.  No other way.
`jsonschema_marshaler_TypeName.go`

First release: only support rendering it in subdirectories.
Second release: support something for elsewhere
*/

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
	case typeregistry.NamedTypeNode, typeregistry.EnumTypeNode, typeregistry.NamedTypeWithAltsNode:
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
	case typeregistry.NamedTypeWithAltsNode:
		return b.renderTypeWithAlts(n)
	case typeregistry.NamedTypeNode:
		child := b.renderChildNode(n.UnderlyingTypeID())
		if n.IsAlt {
			if _, ok := child.(*jsonSchema); !ok {
				panic(fmt.Sprintf("type alt %s must be a struct type", n.NamedType().Obj().String()))
			}
		}
		switch chType := child.(type) {
		case basicMarshaler:
			if rawComments, err := json.Marshal(buildComments(n.TypeSpec.Decorations())); err != nil {
				panic(err)
			} else if len(rawComments) > 0 {
				chType["description"] = json.RawMessage(rawComments)
			}
			return chType
		case *jsonSchema:
			if n.IsAlt {
				discName := n.NamedType().Obj().Name()
				if b.discriminators[discName] {
					for i := 1; ; i++ {
						if !b.discriminators[discName+strconv.Itoa(i)] {
							discName = discName + strconv.Itoa(i)
							break
						}
					}
				}
				b.discriminators[discName] = true
				b.discriminated[n.ID()] = discName
				chType.Properties = append(chType.Properties, schemaProperty{name: b.DiscriminatorPropName, def: constElement(discName)})
			}
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

const propertiesHeader = "\n\n## **Properties**\n\n"

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
		var tempSB strings.Builder
		tempSB.WriteString(propertiesHeader)
		for _, field := range structFieldNodes {
			description := buildComments(field.FieldConf.Field.Decorations())
			if len(description) == 0 {
				continue
			}
			description = strings.TrimPrefix(description, field.FieldConf.Var.Name())
			description = strings.TrimSpace(description)
			tempSB.WriteString("### ")
			tempSB.WriteString(field.FieldConf.FieldName)
			tempSB.WriteString("\n\n")
			tempSB.WriteString(description)
			tempSB.WriteString("\n\n")
		}
		if tempSB.Len() > len(propertiesHeader) {
			sb.WriteString(tempSB.String())
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

// renderTypeWithAlts
//  1. If there is only one child, just render that.
//  2. Otherwise, it has to be given the discriminator const.
//     In order to support this, we have to go back to the TypeRegistry.
//     Named types with multiple alts have to set their type IDS in such a way that
//     they won't be conflated with the same type as a non-alt.
func (b *SchemaBuilder) renderTypeWithAlts(n typeregistry.NamedTypeWithAltsNode) json.Marshaler {
	var (
		alts jsonUnionType
	)

	for _, childID := range n.Children() {
		if schema, ok := b.renderChildNode(childID).(*jsonSchema); ok {
			alts = append(alts, schema)
		} else {
			panic("oh poopoo")
		}
	}
	return alts
}
