package jsonschema_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJsonschema(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Jsonschema Suite")
}
