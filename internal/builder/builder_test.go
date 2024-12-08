package builder

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry"
)

var (
	registry *typeregistry.Registry
)

var _ = BeforeSuite(func() {
	var err error
	registry, err = typeregistry.NewRegistry(nil)
	Expect(err).NotTo(HaveOccurred())
	Expect(registry.LoadAndScan("./testfixtures/...")).To(Succeed())
})

var _ = DescribeTable("Rendering basic types", func(typeName, expected string) {
	//registry.loadAndValidateNamedType()
	ts, ok := registry.GetTypeByName(typeName, mkPkgPath("basic"))

	Expect(ok).To(BeTrue(), "type %s not found", typeName)
	graph, err := registry.GraphTypeForSchema(ts)
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed to resolve graph for %s", typeName))

	b := New(graph)

	result := b.Render()
	data, err := json.Marshal(result)
	Expect(err).NotTo(HaveOccurred(), "failed to marshal result for %s", typeName)
	Expect(data).To(MatchJSON(expected))
},
	Entry("Int type", "Foo", "{\"type\": \"integer\",\"description\": \"Foo is an integer\"}"),
	Entry("Int ptr type", "FooPtr", "{\"type\": \"integer\",\"description\": \"FooPtr is a pointer to int\"}"),
	Entry("String type", "Bar", "{\"type\": \"string\",\"description\": \"Bar is a string\"}"),
	Entry("String ptr type", "BarPtr", "{\"type\": \"string\",\"description\": \"BarPtr is a pointer to string\"}"),
	Entry("Boolean type", "Baz", "{\"type\": \"boolean\",\"description\": \"Baz is a bool\"}"),
	Entry("Boolean ptr type", "BazPtr", "{\"type\": \"boolean\",\"description\": \"BazPtr is a pointer to bool\"}"),
)
