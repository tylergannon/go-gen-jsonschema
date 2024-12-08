package typeregistry

import (
	"errors"
	"fmt"
	"github.com/dave/dst"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go/types"
	"path/filepath"
)

const packageBase = "github.com/tylergannon/go-gen-jsonschema/internal/typeregistry/testfixtures"

type nodesTest = func([]*Node)
type nodeTest = func(*Node)

var _ = Describe("Graphing", func() {

	var (
		registry *Registry
		rootNode *Node
	)

	BeforeEach(func() {
		var err error
		registry, err = NewRegistry(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(registry.LoadAndScan("./testfixtures/testapp0_simple/...")).To(Succeed())
	})
	DescribeTable("Building a type graph", func(typeName string, valid bool, opts ...any) {
		//fmt.Println("Beginning test case ", typeName)
		ts, _, ok := registry.getType(typeName, filepath.Join(packageBase, "testapp0_simple"))
		Expect(ok).To(BeTrue())
		graph, err := registry.GraphTypeForSchema(ts)
		if valid {
			Expect(err).NotTo(HaveOccurred())
			for _, opt := range opts {
				if f, ok := opt.(func(*SchemaGraph)); ok {
					f(graph)
				}
			}
		} else {
			Expect(err).To(HaveOccurred())
		}
	},
		Entry("Declared Type Direct", "DeclaredTypeDirect", true, func(graph *SchemaGraph) {
			Expect(graph.Nodes).To(HaveLen(4))
			Expect(graph.RootNode.ID).To(Equal(TypeID("github.com/tylergannon/go-gen-jsonschema/internal/typeregistry/testfixtures/testapp0_simple.DeclaredTypeDirect")))
		}),
		Entry("Declared Type Pointer", "DeclaredTypePointer", true),
		Entry("Declared Type Definition", "DeclaredTypeDefinition", true),
		Entry("Type Defined As Pointer To Type (definition)", "DeclaredTypeAsPointer", true),
		Entry("Type Defined As type in subpackage", "DeclaredAsRemoteType", true),
		Entry("Declared as slice of remote type", "DeclaredAsSliceOfRemoteType", true),
		Entry("Declared as array of remote type", "DeclaredAsArrayOfRemoteType", true),
		Entry("Struct with embedded types", "StructWithVariousTypes", true),
		Entry("Complicated", "ParentStruct", true, func(graph *SchemaGraph) {
			Expect(graph.Nodes).NotTo(BeEmpty())

			//for k, node := range graph.Nodes {
			//log.Println(k, node.Inbound)
			//}
		}),
	)

	DescribeTable("Checking out the various types", func(typeName string, valid bool, opts ...any) {
		Skip("Wait for it")
		fmt.Println("Beginning test case ", typeName)
		ts, _, ok := registry.getType(typeName, filepath.Join(packageBase, "testapp0_simple"))
		Expect(ok).To(BeTrue())
		id, err := registry.resolveType(ts.GetType(), ts.GetTypeSpec(), ts.pkg)
		Expect(err).NotTo(HaveOccurred())
		rootNode = &Node{
			ID:       id,
			Type:     ts.GetType(),
			TypeSpec: ts,
			Pkg:      ts.Pkg(),
			Node:     ts.GetTypeSpec(),
		}
		nodes, err := registry.visitNode(rootNode)
		if valid {
			Expect(err).NotTo(HaveOccurred())
			for _, opt := range opts {
				if fn, ok := opt.(nodesTest); ok {
					fn(nodes)
				} else if fn, ok := opt.(nodeTest); ok {
					Expect(nodes).To(HaveLen(1))
					fn(nodes[0])
				}
			}
		} else {
			Expect(err).To(HaveOccurred())
		}
	},
		Entry("Declared Type Direct", "DeclaredTypeDirect", true, func(node *Node) {
			Expect(node.Type).To(BeAssignableToTypeOf(&types.Struct{}))
			Expect(node.Node).To(BeAssignableToTypeOf(&dst.StructType{}))

			next, err := registry.visitNode(node)
			Expect(err).NotTo(HaveOccurred())
			Expect(next).To(HaveLen(2))
		}),
		Entry("Declared Type Pointer", "DeclaredTypePointer", true, func(node *Node) {
			Expect(node.Type).To(BeAssignableToTypeOf(&types.Basic{}))
			Expect(node.Node).To(BeAssignableToTypeOf(&dst.Ident{}))
			ident := node.Node.(*dst.Ident)
			Expect(ident.Name).To(Equal("int"))
		}),
		Entry("Declared Type Definition", "DeclaredTypeDefinition", true, func(node *Node) {
			Expect(node.Type).To(BeAssignableToTypeOf(&types.Struct{}))
			Expect(node.Node).To(BeAssignableToTypeOf(&dst.StructType{}))
		}),
		Entry("Type Defined As Pointer To Type (definition)", "DeclaredTypeAsPointer", true, func(node *Node) {
			Expect(node.Type).To(BeAssignableToTypeOf(&types.Named{}))
			Expect(node.Type.(*types.Named).Obj().Name()).To(Equal("DeclaredTypeDefinition"))
			Expect(node.Node).To(BeAssignableToTypeOf(&dst.Ident{}))
			ident := node.Node.(*dst.Ident)
			Expect(ident.Name).To(Equal("DeclaredTypeDefinition"))
		}),
		Entry("Type Defined As type in subpackage", "DeclaredAsRemoteType", true, func(node *Node) {
			Expect(node.Type).To(BeAssignableToTypeOf(&types.Struct{}))
			Expect(node.Node).To(BeAssignableToTypeOf(&dst.StructType{}))
		}),
		Entry("Declared as slice of remote type", "DeclaredAsSliceOfRemoteType", true, func(node *Node) {
			Expect(node.Type).To(BeAssignableToTypeOf(&types.Slice{}))
			Expect(node.Node).To(BeAssignableToTypeOf(&dst.ArrayType{}))
		}),
		Entry("Declared as array of remote type", "DeclaredAsArrayOfRemoteType", true, func(node *Node) {
			Expect(node.Type).To(BeAssignableToTypeOf(&types.Array{}))
			Expect(node.Node).To(BeAssignableToTypeOf(&dst.ArrayType{}))
		}),
		Entry("Struct with embedded types", "StructWithVariousTypes", true, func(node *Node) {
			Expect(node.Type).To(BeAssignableToTypeOf(&types.Struct{}))
			Expect(node.Node).To(BeAssignableToTypeOf(&dst.StructType{}))

			nodes, err := registry.visitNode(node)
			Expect(err).NotTo(HaveOccurred())
			Expect(nodes).To(HaveLen(3))
			Expect(nodes[0].ID).To(Equal(TypeID("int")))
			Expect(nodes[1].ID).To(Equal(TypeID("string")))
			Expect(nodes[2].ID).To(Equal(TypeID("github.com/tylergannon/go-gen-jsonschema/internal/typeregistry/testfixtures/testapp0_simple.DeclaredTypeDirect")))
		}),
	)
	It("Call the graph function", func() {
		Skip("Skip")
		ts, _, ok := registry.getType("SimpleStruct", filepath.Join(packageBase, "testapp0_simple"))
		Expect(ok).To(BeTrue())
		rootNode = &Node{
			Type:     ts.GetType(),
			TypeSpec: ts,
			Pkg:      ts.Pkg(),
			Node:     ts.GetTypeSpec(),
		}
		_, _ = registry.visitNode(rootNode)
	})
})

var _ = Describe("LoadAndValidate", func() {
	var registry *Registry

	BeforeEach(func() {
		var err error
		registry, err = NewRegistry(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(registry.LoadAndScan("./testfixtures/testapp0_simple/...")).To(Succeed())
	})

	DescribeTable("Struct Types", func(typeName string, countFields int, options ...any) {
		ts, _, found := registry.getType(typeName, filepath.Join(packageBase, "testapp0_simple"))
		Expect(found).To(BeTrue(), "item to be found, got %v", found)
		_type := ts.GetType()
		namedType, ok := _type.(*types.Named)
		Expect(ok).To(BeTrue(), "type is not a named type")
		newTypeSpec, err := registry.loadAndValidateNamedType(namedType)
		Expect(err).NotTo(HaveOccurred())
		structType, ok := newTypeSpec.(*NamedStructSpec)
		Expect(ok).To(BeTrue(), "type is not a named type, but %T", newTypeSpec)
		Expect(structType.Fields).To(HaveLen(countFields))
		for _, opt := range options {
			switch _opt := opt.(type) {
			case func([]StructField):
				_opt(structType.Fields)
			}
		}
	},
		Entry("Simple Struct", "SimpleStruct", 4, func(fields []StructField) {
			f0 := fields[0]
			Expect(f0.JSONName).To(Equal("_boolio"))
		}),
		Entry("Simple Struct", "SimpleStructWithPointer", 3, func(fields []StructField) {
			f2 := fields[2]
			Expect(f2.JSONName).To(Equal("baz"))
			Expect(f2.Type).To(BeAssignableToTypeOf(BasicTypeString))
			Expect(f2.Description).To(Equal("There can be multiline comments on a field But in that case, this will be ignored."))
		}),
		Entry("Struct with inline type for field", "EmbeddedStruct", 5, func(fields []StructField) {
			inlineField := fields[2]
			Expect(inlineField.JSONName).To(Equal("__nice__"))
			inlineStruct, ok := inlineField.Type.(*InlineStructSpec)
			Expect(ok).To(BeTrue())
			Expect(inlineStruct.Fields).To(HaveLen(1))
			Expect(inlineStruct.Fields[0].JSONName).To(Equal("node"))
		}),
		Entry("Struct with embedded types", "StructWithEmbeddedField", 6, func(fields []StructField) {
			inlineField := fields[2]
			Expect(inlineField.JSONName).To(Equal("__nice__"))
			inlineStruct, ok := inlineField.Type.(*InlineStructSpec)
			Expect(ok).To(BeTrue())
			Expect(inlineStruct.Fields).To(HaveLen(1))
			Expect(inlineStruct.Fields[0].JSONName).To(Equal("node"))

			Expect(fields[5].JSONName).To(Equal("non_embedded_field"))
		}),
	)
	DescribeTable("Invalid Types", func(typeName string) {
		ts, _, found := registry.getType(typeName, filepath.Join(packageBase, "testapp0_simple"))
		Expect(found).To(BeTrue(), "item to be found, got %v", found)
		_type := ts.GetType()
		namedType, ok := _type.(*types.Named)
		Expect(ok).To(BeTrue(), "type is not a named type")
		_, err := registry.loadAndValidateNamedType(namedType)
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, ErrUnsupportedType)).To(BeTrue(), "expected unsupported type error, got %v", err)
	},
		Entry("Invalid Type (Function Field 1)", "InvalidDueToFunctionField1"),
		Entry("Function Field 2", "InvalidDueToFunctionField2"),
		Entry("Channel Field 1", "InvalidDueToChannelField1"),
		Entry("Channel Field 2", "InvalidDueToChannelField2"),
		Entry("Interface Field 1", "InvalidDueToInterfaceField1"),
		Entry("Interface Field 2", "InvalidDueToInterfaceField2"),
		Entry("Interface Field 3", "InvalidDueToInterfaceField3"),
		Entry("Invalid d/t private field", "InvalidDueToPrivateField"),
	)
})
