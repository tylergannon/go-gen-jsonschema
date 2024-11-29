package typeregistry

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTyperegistry(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Typeregistry Suite")
}
