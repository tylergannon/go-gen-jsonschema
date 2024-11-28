package schemabuilder

import (
	"github.com/dave/dst/decorator"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UnionType Analysis", func() {
	var (
		pkg *decorator.Package
		err error
	)

	BeforeEach(func() {
		// Load the test package containing LLMFriendlyTime
		pkgs, err := decorator.Load(DefaultPackageCfg, "./fixtures/altfindertest/...")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))
		pkg = pkgs[0]
	})

	Context("when analyzing LLMFriendlyTime", func() {
		var info *UnionTypeInfo

		BeforeEach(func() {
			info, err = FindUnionType("LLMFriendlyTime", pkg)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should find all alternatives", func() {
			Expect(info.TypeName).To(Equal("LLMFriendlyTime"))
			Expect(info.Alternatives).To(HaveLen(5))
		})

		It("should correctly identify timeAgo alternative", func() {
			timeAgoAlt := findAltByName(info.Alternatives, "timeAgo")
			Expect(timeAgoAlt).NotTo(BeNil())
			Expect(timeAgoAlt.ConverterType).To(Equal("TimeAgo"))
			Expect(timeAgoAlt.ConverterFunc).To(Equal("ToTime"))
			Expect(timeAgoAlt.ConverterPkg).To(Equal("github.com/tylergannon/go-gen-jsonschema/types"))
		})

		It("should correctly identify timeFromNow alternative", func() {
			timeFromNowAlt := findAltByName(info.Alternatives, "timeFromNow")
			Expect(timeFromNowAlt).NotTo(BeNil())
			Expect(timeFromNowAlt.ConverterType).To(Equal("TimeFromNow"))
			Expect(timeFromNowAlt.ConverterFunc).To(Equal("ToTime"))
			Expect(timeFromNowAlt.ConverterPkg).To(Equal("github.com/tylergannon/go-gen-jsonschema/types"))
		})

		It("should correctly identify actualTime alternative", func() {
			actualTimeAlt := findAltByName(info.Alternatives, "actualTime")
			Expect(actualTimeAlt).NotTo(BeNil())
			Expect(actualTimeAlt.ConverterType).To(Equal("ActualTime"))
			Expect(actualTimeAlt.ConverterFunc).To(Equal("ToTime"))
			Expect(actualTimeAlt.ConverterPkg).To(Equal("github.com/tylergannon/go-gen-jsonschema/types"))
		})

		It("should correctly identify now alternative", func() {
			nowAlt := findAltByName(info.Alternatives, "now")
			Expect(nowAlt).NotTo(BeNil())
			Expect(nowAlt.ConverterType).To(Equal("Now"))
			Expect(nowAlt.ConverterFunc).To(Equal("ToTime"))
			Expect(nowAlt.ConverterPkg).To(Equal("github.com/tylergannon/go-gen-jsonschema/types"))
		})

		It("should correctly identify beginningOfTime alternative", func() {
			botAlt := findAltByName(info.Alternatives, "beginningOfTime")
			Expect(botAlt).NotTo(BeNil())
			Expect(botAlt.ConverterType).To(Equal("BeginningOfTime"))
			Expect(botAlt.ConverterFunc).To(Equal("ToTime"))
			Expect(botAlt.ConverterPkg).To(Equal("github.com/tylergannon/go-gen-jsonschema/types"))
		})
	})

	Context("when analyzing a non-existent type", func() {
		It("should return an error", func() {
			_, err := FindUnionType("NonExistentType", pkg)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no union type definition found"))
		})
	})
})

// Helper function to find an alternative by name
func findAltByName(alts []AltFunction, name string) *AltFunction {
	for _, alt := range alts {
		if alt.Name == name {
			return &alt
		}
	}
	return nil
}
