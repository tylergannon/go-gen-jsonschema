package typeregistry

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"path/filepath"
)

var _ = Describe("Type Alternatives", func() {
	var (
		registry *Registry
	)

	BeforeEach(func() {
		var err error
		registry, err = NewRegistry(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(registry.LoadAndScan("./testfixtures/registrytestapp/...")).To(Succeed())
	})
	When("There are type alts", func() {
		It("Should find them", func() {
			ts, ok := registry.getType("LLMFriendlyTime", filepath.Join(packageBase, "registrytestapp"))
			Expect(ok).To(BeTrue())
			graph, err := registry.GraphTypeForSchema(ts)
			Expect(err).NotTo(HaveOccurred())
			Expect(graph.RootNode).To(BeAssignableToTypeOf(NamedTypeWithAltsNode{}))
			node := graph.RootNode.(NamedTypeWithAltsNode)
			Expect(node.TypeSpec.Alternatives()).To(HaveLen(5))
			alts := node.TypeSpec.Alternatives()
			Expect(alts[0].Alias).To(Equal("timeAgo"))
			Expect(alts[1].Alias).To(Equal("timeFromNow"))
			Expect(alts[2].Alias).To(Equal("actualTime"))
			Expect(alts[3].Alias).To(Equal("now"))
			Expect(alts[4].Alias).To(Equal("beginningOfTime"))
			Expect(node.nodeImpl.children).To(HaveLen(5))
			ch := node.nodeImpl.children
			Expect(ch[0]).To(Equal(TypeID("github.com/tylergannon/go-gen-jsonschema/internal/typeregistry/testfixtures/registrytestapp/subpkg.TimeAgo~")))
			Expect(ch[1]).To(Equal(TypeID("github.com/tylergannon/go-gen-jsonschema/internal/typeregistry/testfixtures/registrytestapp/subpkg.TimeFromNow~")))
			Expect(ch[2]).To(Equal(TypeID("github.com/tylergannon/go-gen-jsonschema/internal/typeregistry/testfixtures/registrytestapp/subpkg.ActualTime~")))
			Expect(ch[3]).To(Equal(TypeID("github.com/tylergannon/go-gen-jsonschema/internal/typeregistry/testfixtures/registrytestapp/subpkg.Now~")))
			Expect(ch[4]).To(Equal(TypeID("github.com/tylergannon/go-gen-jsonschema/internal/typeregistry/testfixtures/registrytestapp/subpkg.BeginningOfTime~")))
			_ = graph
		})
	})
})
