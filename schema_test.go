package jsonschema

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JSONSchema with Strict field", func() {
	var schema = JSONSchema{
		Type:        Object,
		Description: "Test schema",
		Properties: map[string]json.Marshaler{
			"foo": BoolSchema("bar"),
			"bax": StringSchema("quux"),
		},
		Strict: true,
	}

	It("Sets additionalProperties to false when Strict is true", func() {
		data, err := json.Marshal(schema)
		Expect(err).NotTo(HaveOccurred())

		// Unmarshal to verify the output
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		Expect(err).NotTo(HaveOccurred())

		// Check that additionalProperties is set to false
		Expect(result).To(HaveKey("additionalProperties"))
		Expect(result["additionalProperties"]).To(BeFalse())
	})

	It("Includes all property keys in the required array when Strict is true", func() {
		data, err := json.Marshal(schema)
		Expect(err).NotTo(HaveOccurred())

		// Unmarshal to verify the output
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		Expect(err).NotTo(HaveOccurred())

		// Check that required contains all property keys
		Expect(result).To(HaveKey("required"))
		required, ok := result["required"].([]interface{})
		Expect(ok).To(BeTrue())
		Expect(required).To(HaveLen(2))
		Expect(required).To(ContainElements("foo", "bax"))
	})

	It("Does not include the Strict field in the marshaled output", func() {
		data, err := json.Marshal(schema)
		Expect(err).NotTo(HaveOccurred())

		// Unmarshal to verify the output
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		Expect(err).NotTo(HaveOccurred())

		// Strict should not be in the output
		Expect(result).NotTo(HaveKey("Strict"))
		Expect(result).NotTo(HaveKey("strict"))
	})
})
