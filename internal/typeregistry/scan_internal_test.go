package typeregistry

import (
	"github.com/dave/dst"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go/types"
	"path/filepath"
)

const packageBase = "github.com/tylergannon/go-gen-jsonschema/internal/typeregistry/testfixtures"

type nodesTest = func([]nodeInternal)
type nodeTest = func(internal nodeInternal)

var _ = Describe("Graphing", func() {

	var (
		registry *Registry
		rootNode NamedTypeNode
	)

	BeforeEach(func() {
		var err error
		registry, err = NewRegistry(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(registry.LoadAndScan("./testfixtures/testapp0_simple/...")).To(Succeed())
	})
	DescribeTable("Building a type graph", func(typeName string, valid bool, opts ...any) {
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
			Expect(graph.Nodes).To(HaveLen(6))
			Expect(graph.RootNode.ID()).To(Equal(TypeID("github.com/tylergannon/go-gen-jsonschema/internal/typeregistry/testfixtures/testapp0_simple.DeclaredTypeDirect")))
		}),
		Entry("Declared Type Pointer", "DeclaredTypePointer", true),
		Entry("Declared Type Definition", "DeclaredTypeDefinition", true),
		Entry("Type Defined As Pointer To Type (definition)", "DeclaredTypeAsPointer", true),
		Entry("Type Defined As type in subpackage", "DeclaredAsRemoteType", true),
		Entry("Declared as slice of remote type", "DeclaredAsSliceOfRemoteType", true),
		//Entry("Declared as array of remote type", "DeclaredAsArrayOfRemoteType", true),
		Entry("Struct with embedded types", "StructWithVariousTypes", true),
		Entry("Complicated", "ParentStruct", true, func(graph *SchemaGraph) {
			Expect(graph.Nodes).NotTo(BeEmpty())

			//for k, node := range graph.Nodes {
			//log.Println(k, node.Inbound)
			//}
		}),
	)

	DescribeTable("Checking out the various types", func(typeName string, valid bool, opts ...any) {
		ts, _, ok := registry.getType(typeName, filepath.Join(packageBase, "testapp0_simple"))
		Expect(ok).To(BeTrue())
		id, err := registry.resolveType(ts.GetType(), ts.GetTypeSpec(), ts.pkg)
		Expect(err).NotTo(HaveOccurred())
		rootNode = NamedTypeNode{TypeSpec: ts, nodeImpl: &nodeImpl{
			id:      id,
			typ:     ts.GetType(),
			pkg:     ts.Pkg(),
			dstNode: ts.GetTypeSpec(),
		}}
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
		Entry("Declared Type Direct", "DeclaredTypeDirect", true, func(node nodeInternal) {
			Expect(node.Type()).To(BeAssignableToTypeOf(&types.Struct{}))
			Expect(node.DSTNode()).To(BeAssignableToTypeOf(&dst.StructType{}))

			next, err := registry.visitNode(node)
			Expect(err).NotTo(HaveOccurred())
			Expect(next).To(HaveLen(2))
		}),
		Entry("Declared Type Pointer", "DeclaredTypePointer", true, func(node nodeInternal) {
			Expect(node.Type()).To(BeAssignableToTypeOf(&types.Basic{}))
			Expect(node.DSTNode()).To(BeAssignableToTypeOf(&dst.Ident{}))
			ident := node.DSTNode().(*dst.Ident)
			Expect(ident.Name).To(Equal("int"))
		}),
		Entry("Declared Type Definition", "DeclaredTypeDefinition", true, func(node nodeInternal) {
			Expect(node.Type()).To(BeAssignableToTypeOf(&types.Struct{}))
			Expect(node.DSTNode()).To(BeAssignableToTypeOf(&dst.StructType{}))
		}),
		Entry("Type Defined As Pointer To Type (definition)", "DeclaredTypeAsPointer", true, func(node nodeInternal) {
			Expect(node.Type()).To(BeAssignableToTypeOf(&types.Named{}))
			Expect(node.Type().(*types.Named).Obj().Name()).To(Equal("DeclaredTypeDefinition"))
			Expect(node.DSTNode()).To(BeAssignableToTypeOf(&dst.TypeSpec{}))
			ident := node.DSTNode().(*dst.TypeSpec)
			Expect(ident.Name.Name).To(Equal("DeclaredTypeDefinition"))
		}),
		Entry("Type Defined As type in subpackage", "DeclaredAsRemoteType", true, func(node nodeInternal) {
			Expect(node.Type()).To(BeAssignableToTypeOf(&types.Struct{}))
			Expect(node.DSTNode()).To(BeAssignableToTypeOf(&dst.StructType{}))
		}),
		Entry("Declared as slice of remote type", "DeclaredAsSliceOfRemoteType", true, func(node nodeInternal) {
			Expect(node.Type()).To(BeAssignableToTypeOf(&types.Slice{}))
			Expect(node.DSTNode()).To(BeAssignableToTypeOf(&dst.ArrayType{}))
		}),
		Entry("Declared as array of remote type", "DeclaredAsArrayOfRemoteType", true, func(node nodeInternal) {
			Expect(node.Type()).To(BeAssignableToTypeOf(&types.Array{}))
			Expect(node.DSTNode()).To(BeAssignableToTypeOf(&dst.ArrayType{}))
		}),
		Entry("Struct with embedded types", "StructWithVariousTypes", true, func(node nodeInternal) {
			Expect(node.Type()).To(BeAssignableToTypeOf(&types.Struct{}))
			Expect(node.DSTNode()).To(BeAssignableToTypeOf(&dst.StructType{}))

			nodes, err := registry.visitNode(node)
			Expect(err).NotTo(HaveOccurred())
			Expect(nodes).To(HaveLen(3))
			Expect(nodes[0].ID()).To(Equal(TypeID("struct {Foo int; Bar string; Field2 ../testapp0_simple.DeclaredTypeDirect;}!Foo")))
			Expect(nodes[1].ID()).To(Equal(TypeID("struct {Foo int; Bar string; Field2 ../testapp0_simple.DeclaredTypeDirect;}!Bar")))
			Expect(nodes[2].ID()).To(Equal(TypeID("struct {Foo int; Bar string; Field2 ../testapp0_simple.DeclaredTypeDirect;}!Field2")))
		}),
	)
	It("Call the graph function", func() {
		Skip("Skip")
		ts, _, ok := registry.getType("SimpleStruct", filepath.Join(packageBase, "testapp0_simple"))
		Expect(ok).To(BeTrue())
		rootNode = NamedTypeNode{
			TypeSpec: ts,
			nodeImpl: &nodeImpl{
				typ:     ts.GetType(),
				pkg:     ts.Pkg(),
				dstNode: ts.GetTypeSpec(),
			},
		}
		_, _ = registry.visitNode(rootNode)
	})
})
