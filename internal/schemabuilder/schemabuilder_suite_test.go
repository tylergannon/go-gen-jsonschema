package schemabuilder

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSchemabuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Schemabuilder Suite")
}
