package typeregistry

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Type Resolution", func() {
	const (
		appName         = "testapp1_typeresolution"
		pkgRelativePath = "./testfixtures/testapp1_typeresolution/..."
	)
	var (
		registry *Registry
		pkgPath  = fmt.Sprintf("%s/%s", packageBase, appName)
	)

	BeforeEach(func() {
		var err error
		registry, err = NewRegistry(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(registry.LoadAndScan(pkgRelativePath)).To(Succeed())
	})
	When("Given a named type", func() {
		It("returns the canonical name for that type", func() {
			var ts, found = registry.getType("TrivialNamedType", pkgPath)
			Expect(found).To(BeTrue())
			actual, err := registry.resolveType(ts.GetType(), ts.typeSpec, ts.pkg)
			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal(NewTypeID(pkgPath, "TrivialNamedType")))
		})
	})

	It("Finds something foo", func() {
		//var ts, _, found = registry.getType("ArrayType", pkgPath)
		//Expect(found).To(BeTrue())
		//inspect("underlying: ", ts.GetType().(*types.Named).Underlying(), ts.TypeSpec.Type)
	})

	var typeName = func(indexer, typeName string) string {
		return fmt.Sprintf("%s.%s%s", pkgPath, typeName, indexer)
	}

	type testFunc func(t TypeID, e error)

	DescribeTable("Checking inline and complex types", func(typeName, expected string, success bool, tests ...testFunc) {
		var (
			ts, found = registry.getType(typeName, pkgPath)
		)
		Expect(found).To(BeTrue(), fmt.Sprintf("type %s not found", typeName))
		actual, err := registry.resolveType(ts.GetType().Underlying(), ts.typeSpec.Type, ts.pkg)
		if success {
			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal(TypeID(expected)))
		} else {
			Expect(err).To(HaveOccurred())
		}
		for _, test := range tests {
			test(actual, err)
		}
	},
		Entry("Basic type int", "TrivialNamedType", "int", true),
		Entry("Basic type string", "TrivialNamedTypeString", "string", true),
		Entry("Pointers not represented", "TrivialNamedTypePtrInt", "int", true),
		Entry("Pointers not represented (string)", "TrivialNamedTypePtrString", "string", true),
		Entry("ArrayType", "ArrayType", typeName("[3]", "TrivialNamedType"), true),
		Entry("SliceType", "SliceType", typeName("[]", "TrivialNamedType"), true),
		Entry("Slice of Slice Type", "SliceOfSlice", typeName("[][]", "TrivialNamedType"), true),

		Entry("Pointers ArrayType", "PointersArrayType", typeName("[3]", "TrivialNamedType"), true),
		Entry("Pointers SliceType", "PointersSliceType", typeName("[]", "TrivialNamedType"), true),
		Entry("Pointers Slice of Slice Type", "PointersSliceOfSlice", typeName("[][]", "TrivialNamedType"), true),

		Entry("Struct Type Simple", "StructTypeSimple", "struct {}", true),
		Entry("Struct Type With Attrs", "StructTypeWithStuff", "struct {Foo string; Bar int;}", true),
		Entry("Struct Type With Attrs (ptrs)", "PointersStructTypeWithStuff", "struct {Foo string; Bar int;}", true),
		Entry("Struct Type With Attrs (ptrs, json tags)", "PointersStructTypeWithJsonTabs", "struct {foo string `json:\"foo,omitempty\"`; bar int `json:\"bar,omitempty\"`;}", true),
		Entry("Struct Type With Attrs (ptrs, json tags) With IgnoreFields", "PointersStructTypeWithIgnoreFields", "struct {foo string `json:\"foo,omitempty\"`; bar int `json:\"bar,omitempty\"`;}", true),
		Entry("Struct type with embedded struct", "StructWithEmbeddings", "struct {Bat string; Foo string; Bar int;}", true),
		Entry("Struct type with inline struct", "StructWithInline", "struct {Foo int; Bar struct {Foo int; Bar string;};}", true),
		Entry("Struct type with inline struct and named types", "StructWithInlineAndNamed", "struct {Foo ../testapp1_typeresolution.ArrayType; Bar struct {Foo ../testapp1_typeresolution.SliceOfSlice; Bar ../testapp1_typeresolution.SliceOfSlice[10];};}", true),
		Entry("Crazy Struct type with inline struct and named types", "StructWithInlineAndNamedAllCrazy", "struct {Foo ../testapp1_typeresolution.ArrayType; Bar struct {Foo ../testapp1_typeresolution.SliceOfSlice; Bar ../testapp1_typeresolution.SliceOfSlice[10]; __nobody__ struct {Bark ../testapp1_typeresolution.SliceOfSlice; Bite int; Recurse ../testapp1_typeresolution.StructWithInlineAndNamedAllCrazy; Foo ../testapp1_typeresolution.ArrayType; Bar struct {Foo ../testapp1_typeresolution.SliceOfSlice; Bar ../testapp1_typeresolution.SliceOfSlice[10];}; boop struct {bat int `json:\"bat\"`;} `json:\"boop\"`;}[][] `json:\"__nobody__\"`; Foo ../testapp1_typeresolution.ArrayType; Bar struct {Foo ../testapp1_typeresolution.SliceOfSlice; Bar ../testapp1_typeresolution.SliceOfSlice[10];};};}", true),

		Entry("Crazy Recursive", "ParentStruct", "struct {Inline struct {Bar int; Baz string; Coolio bool; Child ../testapp1_typeresolution.ChildStruct;}; Foobar ../testapp1_typeresolution.ParentStruct; Inline struct {Bar int; Bark string; Coolio bool; Foobar ../testapp1_typeresolution.ChildStruct; Inline struct {Bar int;};}; GoodKid ../testapp1_typeresolution.ChildStruct; BadKid ../testapp1_typeresolution.ChildStruct;}", true),

		// Fail
		Entry("Crazy struct with a channel in", "IllegalStructWithInlineAndNamedAllCrazy", "", false, func(_ TypeID, e error) {
			//Expect(e).To(HaveOccurred())
			//fmt.Println(e)
		}),
		Entry("Crazy Recursive", "ParentStructRecursive", "", false),
		Entry("Interface Type", "SomeInterface", "", false, func(_ TypeID, e error) {
			Expect(e).To(HaveOccurred())
			fmt.Println(e)
		}),
		Entry("Struct With Interface Field", "StructWithInterfaceField", "", false),
		Entry("Struct With Embedded Interface", "StructWithEmbeddedInterface", "", false),
	)
})
