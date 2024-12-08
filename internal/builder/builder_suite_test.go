package builder

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Builder Suite")
}

type literalMarshaler []byte

func (l literalMarshaler) MarshalJSON() ([]byte, error) {
	return l, nil
}

var _ = Describe("jsonSchema", func() {
	It("marshals with strict = true and multiple properties", func() {
		s := jsonSchema{
			Description: "A strict schema",
			Properties: []schemaProperty{
				{name: "one", def: literalMarshaler(`{"type":"string","minLength":1}`)},
				{name: "two", def: literalMarshaler(`{"type":"number"}`)},
				{name: "three", def: literalMarshaler(`{"type":"boolean"}`)},
			},
			Strict: true,
		}
		data, err := json.Marshal(&s)
		Expect(err).To(BeNil())
		Expect(data).To(MatchJSON(`{
			"description":"A strict schema",
			"type":"object",
			"properties":{
				"one":{"type":"string","minLength":1},
				"two":{"type":"number"},
				"three":{"type":"boolean"}
			},
			"required":["one","two","three"],
			"additionalProperties":false
		}`))
	})

	It("marshals with strict = false, no required, but definitions ($refs)", func() {
		s := jsonSchema{
			Description: "Schema with definitions",
			Properties: []schemaProperty{
				{name: "alpha", def: literalMarshaler(`{"type":"string"}`)},
			},
			Definitions: map[string]json.Marshaler{
				"RefType": literalMarshaler(`{"type":"object","properties":{"id":{"type":"string"}}}`),
			},
			DefinitionsKey: "$defs",
		}
		data, err := json.Marshal(&s)
		Expect(err).To(BeNil())
		Expect(data).To(MatchJSON(`{
			"description":"Schema with definitions",
			"type":"object",
			"properties":{"alpha":{"type":"string"}},
			"$defs":{
				"RefType":{"type":"object","properties":{"id":{"type":"string"}}}
			}
		}`))
	})

	It("marshals with strict = false, no required, but definitions", func() {
		s := jsonSchema{
			Description: "Schema with definitions",
			Properties: []schemaProperty{
				{name: "alpha", def: literalMarshaler(`{"type":"string"}`)},
			},
			Definitions: map[string]json.Marshaler{
				"RefType": literalMarshaler(`{"type":"object","properties":{"id":{"type":"string"}}}`),
			},
		}
		data, err := json.Marshal(&s)
		Expect(err).To(BeNil())
		Expect(data).To(MatchJSON(`{
			"description":"Schema with definitions",
			"type":"object",
			"properties":{"alpha":{"type":"string"}},
			"definitions":{
				"RefType":{"type":"object","properties":{"id":{"type":"string"}}}
			}
		}`))
	})

	It("marshals with strict = false, explicit required, and no definitions", func() {
		s := jsonSchema{
			Description: "Schema with explicit required",
			Properties: []schemaProperty{
				{name: "propA", def: literalMarshaler(`{"type":"string"}`)},
				{name: "propB", def: literalMarshaler(`{"type":"integer"}`)},
			},
			Required: []string{"propA"},
		}
		data, err := json.Marshal(&s)
		Expect(err).To(BeNil())
		Expect(data).To(MatchJSON(`{
			"description":"Schema with explicit required",
			"type":"object",
			"properties":{
				"propA":{"type":"string"},
				"propB":{"type":"integer"}
			},
			"required":["propA"]
		}`))
	})

	It("marshals with strict = false, additionalProperties as literal object", func() {
		s := jsonSchema{
			Description: "Schema with additionalProperties",
			Properties: []schemaProperty{
				{name: "propX", def: literalMarshaler(`{"type":"string"}`)},
			},
			AdditionalProperties: literalMarshaler(`{"type":"string"}`),
		}
		data, err := json.Marshal(&s)
		Expect(err).To(BeNil())
		Expect(data).To(MatchJSON(`{
			"description":"Schema with additionalProperties",
			"type":"object",
			"properties":{"propX":{"type":"string"}},
			"additionalProperties":{"type":"string"}
		}`))
	})

	It("marshals with no properties", func() {
		s := jsonSchema{
			Description: "Empty object schema",
		}
		data, err := json.Marshal(&s)
		Expect(err).To(BeNil())
		Expect(data).To(MatchJSON(`{
			"description":"Empty object schema",
			"type":"object"
		}`))
	})
})
