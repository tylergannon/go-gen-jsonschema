package typeregistry

import (
	"errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go/types"
	"path/filepath"
)

const packageBase = "github.com/tylergannon/go-gen-jsonschema/internal/typeregistry/testfixtures"

//
//var _ = Describe("scan", func() {
//	var (
//		pkgs     []*decorator.Package
//		registry *Registry
//	)
//	BeforeEach(func() {
//		var err error
//		pkgs, err = loader.Load("./testfixtures/registrytestapp/...")
//		Expect(err).NotTo(HaveOccurred())
//		registry, err = NewRegistry(pkgs)
//		Expect(err).NotTo(HaveOccurred())
//	})
//	It("Should all be all right ", func() {
//		Expect(true)
//		Expect(registry.unionTypes).To(HaveLen(2))
//	})
//})

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
		newTypeSpec, err := registry.LoadAndValidateNamedType(namedType)
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
		_, err := registry.LoadAndValidateNamedType(namedType)
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
