package schemabuilder

import (
	"github.com/tylergannon/go-gen-jsonschema/internal/loader"
	"github.com/tylergannon/go-gen-jsonschema/internal/typeregistry"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSchemabuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Schemabuilder Suite")
}

const (
	fixturePath      = "./fixtures"
	localPackagePath = "github.com/tylergannon/go-gen-jsonschema/internal/schemabuilder"
)

func loadTestData(fixtureName, typeName string) *SchemaBuilder {
	pkgs, err := loader.Load(filepath.Join(fixturePath, fixtureName, "..."))
	Expect(err).NotTo(HaveOccurred())
	packagePath := filepath.Join(localPackagePath, fixturePath, fixtureName)
	registry, err := typeregistry.NewRegistry(pkgs)
	Expect(err).NotTo(HaveOccurred())
	builder, err := New(typeName, packagePath, registry)
	Expect(err).NotTo(HaveOccurred())
	return builder
}

func loadTestFile(fixtureName, typeName string) []byte {
	result, err := os.ReadFile(filepath.Join(fixturePath, fixtureName, typeName+".expected.json"))
	Expect(err).NotTo(HaveOccurred())
	return result
}

func writeTestFile(fixtureName, typeName string, data []byte) {
	Expect(os.WriteFile(filepath.Join(fixturePath, fixtureName, typeName+".expected.json"), data, 0644)).To(Succeed())
}
