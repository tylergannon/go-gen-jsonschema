package schemabuilder

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/tools/go/packages"
)

var _ = Describe("StructFinder", func() {
	var pkg *packages.Package
	BeforeEach(func() {
		pkgs, err := packages.Load(DefaultPackageCfg, "./fixtures/testapp1")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))
		pkg = pkgs[0]
	})
	It("Finds the struct", func() {
		info, err := findTypeInfoInPackage(pkg, "SimpleStruct")
		Expect(err).NotTo(HaveOccurred())
		Expect(info.named.Obj().Name()).To(Equal("SimpleStruct"))
		Expect(info.comments).To(Equal("Build this struct in order to really get a lot of meaning out of life.\nIt's really essential that you get all of this down."))
	})
})
