package jsonschema

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ObjectSchema", func() {
	Context("When marshaling to JSON", func() {
		It("produces a valid JSON Schema with type:object", func() {
			schema := &ObjectSchema{
				Description: "Test schema",
			}

			data, err := json.Marshal(schema)
			Expect(err).NotTo(HaveOccurred())

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(HaveKey("type"))
			Expect(result["type"]).To(Equal("object"))
		})

		It("includes properties when they are provided", func() {
			schema := &ObjectSchema{
				Description: "Test schema",
			}

			schema.AddProperty("foo", BoolSchema("A boolean property"))
			schema.AddProperty("bar", StringSchema("A string property"))

			data, err := json.Marshal(schema)
			Expect(err).NotTo(HaveOccurred())

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(HaveKey("properties"))
			props, ok := result["properties"].(map[string]interface{})
			Expect(ok).To(BeTrue())

			Expect(props).To(HaveKey("foo"))
			Expect(props).To(HaveKey("bar"))

			foo := props["foo"].(map[string]interface{})
			Expect(foo["type"]).To(Equal("boolean"))
			Expect(foo["description"]).To(Equal("A boolean property"))

			bar := props["bar"].(map[string]interface{})
			Expect(bar["type"]).To(Equal("string"))
			Expect(bar["description"]).To(Equal("A string property"))
		})

		It("includes required properties when they are added", func() {
			schema := &ObjectSchema{
				Description: "Test schema",
			}

			schema.AddProperty("foo", BoolSchema("A boolean property"))
			schema.AddRequiredProperty("bar", StringSchema("A required string property"))

			data, err := json.Marshal(schema)
			Expect(err).NotTo(HaveOccurred())

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(HaveKey("required"))
			required, ok := result["required"].([]interface{})
			Expect(ok).To(BeTrue())
			Expect(required).To(HaveLen(1))
			Expect(required[0]).To(Equal("bar"))
		})

		It("includes the description when provided", func() {
			schema := &ObjectSchema{
				Description: "Test schema description",
			}

			data, err := json.Marshal(schema)
			Expect(err).NotTo(HaveOccurred())

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(HaveKey("description"))
			Expect(result["description"]).To(Equal("Test schema description"))
		})

		It("includes additionalProperties when provided", func() {
			schema := &ObjectSchema{
				Description:          "Test schema",
				AdditionalProperties: true,
			}

			data, err := json.Marshal(schema)
			Expect(err).NotTo(HaveOccurred())

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(HaveKey("additionalProperties"))
			Expect(result["additionalProperties"]).To(BeTrue())
		})

		Context("With Strict = true", func() {
			It("makes all properties required and additionalProperties false", func() {
				schema := &ObjectSchema{
					Description: "Test schema with strict mode",
					Strict:      true,
				}

				schema.AddProperty("foo", BoolSchema("A boolean property"))
				schema.AddProperty("bar", StringSchema("A string property"))

				data, err := json.Marshal(schema)
				Expect(err).NotTo(HaveOccurred())

				Expect(data).To(Equal([]byte(`{"type":"object","description":"Test schema with strict mode","properties":{"foo":{"description":"A boolean property","type":"boolean"},"bar":{"description":"A string property","type":"string"}},"required":["foo","bar"],"additionalProperties":false}`)))

				var result map[string]interface{}
				err = json.Unmarshal(data, &result)
				Expect(err).NotTo(HaveOccurred())

				// Check additionalProperties is false
				Expect(result).To(HaveKey("additionalProperties"))
				Expect(result["additionalProperties"]).To(BeFalse())

				// Check all properties are required
				Expect(result).To(HaveKey("required"))
				required, ok := result["required"].([]interface{})
				Expect(ok).To(BeTrue())
				Expect(required).To(HaveLen(2))
				Expect(required).To(ContainElements("foo", "bar"))
			})

			It("overrides explicit additionalProperties when Strict is true", func() {
				schema := &ObjectSchema{
					Description:          "Test schema with strict mode",
					Strict:               true,
					AdditionalProperties: true, // This should be overridden
				}

				schema.AddProperty("foo", BoolSchema("A boolean property"))

				data, err := json.Marshal(schema)
				Expect(err).NotTo(HaveOccurred())

				var result map[string]interface{}
				err = json.Unmarshal(data, &result)
				Expect(err).NotTo(HaveOccurred())

				// Check additionalProperties is still false because Strict is true
				Expect(result).To(HaveKey("additionalProperties"))
				Expect(result["additionalProperties"]).To(BeFalse())
			})

			It("overrides explicit required properties when Strict is true", func() {
				schema := &ObjectSchema{
					Description: "Test schema with strict mode",
					Strict:      true,
					Required:    []string{"foo"}, // This will be overridden
				}

				schema.AddProperty("foo", BoolSchema("A boolean property"))
				schema.AddProperty("bar", StringSchema("A string property"))

				data, err := json.Marshal(schema)
				Expect(err).NotTo(HaveOccurred())

				var result map[string]interface{}
				err = json.Unmarshal(data, &result)
				Expect(err).NotTo(HaveOccurred())

				// Check all properties are required, not just "foo"
				Expect(result).To(HaveKey("required"))
				required, ok := result["required"].([]interface{})
				Expect(ok).To(BeTrue())
				Expect(required).To(HaveLen(2))
				Expect(required).To(ContainElements("foo", "bar"))
			})
		})

		It("does not include the Strict field in the marshaled output", func() {
			schema := &ObjectSchema{
				Description: "Test schema",
				Strict:      true,
			}

			data, err := json.Marshal(schema)
			Expect(err).NotTo(HaveOccurred())

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			Expect(err).NotTo(HaveOccurred())

			// Strict should not be in the output
			Expect(result).NotTo(HaveKey("Strict"))
			Expect(result).NotTo(HaveKey("strict"))
		})
	})
})
