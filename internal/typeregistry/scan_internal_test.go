package typeregistry

import (
	"github.com/dave/dst/decorator"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tylergannon/go-gen-jsonschema/internal/schemabuilder"
)

var _ = Describe("Scan", func() {
	var (
		pkgs     []*decorator.Package
		registry *Registry
	)
	BeforeEach(func() {
		var err error
		pkgs, err = decorator.Load(schemabuilder.DefaultPackageCfg, "./testfixtures/registrytestapp/...")
		Expect(err).NotTo(HaveOccurred())
		registry, err = NewRegistry(pkgs)
		Expect(err).NotTo(HaveOccurred())
	})
	It("Should all be all right ", func() {
		Expect(true)
		Expect(registry.unionTypes).To(HaveLen(3))
	})
})
