package schemabuilder

import (
	"github.com/dave/dst/decorator"
	"slices"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AST Struct Resolution", func() {
	var (
		importMap *PackageMap
	)
	BeforeEach(func() {
		var err error

		pkgs, err := decorator.Load(DefaultPackageCfg, "./fixtures/testapp1/...")
		Expect(err).To(BeNil())

		Expect(pkgs).ToNot(BeEmpty())
		mainPkgIdx := slices.IndexFunc(pkgs, func(p *decorator.Package) bool {
			return p.Name == "testapp1"
		})
		Expect(mainPkgIdx).To(BeNumerically(">=", 0))
		importMap = NewPackageMap(pkgs[mainPkgIdx])
		for _, pkg := range pkgs {
			importMap.AddPackage(pkg)
		}
	})

	DescribeTable("should resolve struct definitions and field comments", func(fieldName string, preComments string) {
		structType, err := resolveStruct(importMap, fieldName, importMap.localPackage.PkgPath)
		Expect(err).To(BeNil())

		comments := strings.Join(structType.TypeSpec.Decorations().Start.All(), "")
		Expect(comments).To(Equal(preComments))

		fieldComments := extractFieldComments(structType.StructType)
		Expect(fieldComments).To(HaveKeyWithValue("Foo", "There can be comments here"))
		Expect(fieldComments).To(HaveKeyWithValue("Bar", "There can also be comments to the right"))
		Expect(fieldComments).To(HaveKeyWithValue("Baz", "There can be\nmultiline comments\non a field\nBut in that case, this will be ignored."))

		//Expect(fieldComments).To(HaveKeyWithValue("DefinedElsewhere", ""))
		Expect(fieldComments).To(HaveKey("DefinedElsewhere"))

	},
		Entry("Actual Definition defined in block", "ComplexExample", "// Build this struct in order to really get a lot of meaning out of life.// It's really essential that you get all of this down."),
		Entry("Type Definition defined in block", "ComplexDefinition", "// These are the comments that will be used, not the ones on the other type."),
		Entry("Type Alias defined in block", "ComplexAlias", "// These comments will be used, not the ones on the aliased type."),
		Entry("Actual Definition defined in single GenDecl", "ComplexExample2", "// Build this struct in order to really get a lot of meaning out of life.// It's really essential that you get all of this down."),
		Entry("Type Definition defined in single GenDecl", "ComplexDefinition2", "// These are the comments that will be used, not the ones on the other type."),
		Entry("Type Alias defined in single GenDecl", "ComplexAlias2", "// These comments will be used, not the ones on the aliased type."),
	)
	DescribeTable("resolving struct definitions from another package", func(fieldName string, comment string) {
		structType, err := resolveStruct(importMap, fieldName, importMap.localPackage.PkgPath)
		Expect(err).To(BeNil())

		comments := strings.Join(structType.TypeSpec.Decorations().Start.All(), "")
		Expect(comments).To(Equal(comment))
		fieldComments := extractFieldComments(structType.StructType)
		Expect(fieldComments).To(HaveKeyWithValue("Attr1", "Excellent"))
		Expect(fieldComments).To(HaveKeyWithValue("Attr2", "Very Cool"))
	},
		Entry("Remote Definition in block", "RemoteDefinition", "// These are the comments that will be used, not the ones on the other type."),
		Entry("Remote Alias in block", "RemoteAlias", "// These comments will be used, not the ones on the aliased type."),
		Entry("Remote Definition in single GenDecl", "RemoteDefinition2", "// These are the comments that will be used, not the ones on the other type."),
		Entry("Remote Alias in single GenDecl", "RemoteAlias2", "// These comments will be used, not the ones on the aliased type."),
	)
})
