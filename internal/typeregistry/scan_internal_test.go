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
	DescribeTable("Invalid Types", func(typeName string) {
		ts, _, found := registry.GetType(typeName, filepath.Join(packageBase, "testapp0_simple"))
		Expect(found).To(BeTrue(), "item to be found, got %v", found)
		_type := ts.GetType().Type()
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
