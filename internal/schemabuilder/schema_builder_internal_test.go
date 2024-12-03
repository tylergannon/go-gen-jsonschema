package schemabuilder

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const record = false

var _ = Describe("SchemaBuilder", func() {
	//Describe("Simple Struct", func() {
	//	var builder *SchemaBuilder
	//	BeforeEach(func() {
	//		builder = loadTestData("testapp0_simple", "SimpleStruct")
	//	})
	//	It("Has the basic type in it", func() {
	//		_, err := json.MarshalIndent(builder, "", "  ")
	//		Expect(err).ToNot(HaveOccurred())
	//		//log.Println(string(data))
	//	})
	//})
	//Describe("Simple Struct With Pointer", func() {
	//	var builder *SchemaBuilder
	//	BeforeEach(func() {
	//		builder = loadTestData("testapp0_simple", "SimpleStructWithPointer")
	//	})
	//	It("Has the basic type in it", func() {
	//		data, err := json.MarshalIndent(builder, "", "  ")
	//		Expect(err).ToNot(HaveOccurred())
	//		log.Println(string(data))
	//	})
	//})
	DescribeTable("Schemas Are Produced Correctly", func(fixtureName, typeName string, recordThisOne ...bool) {
		builder := loadTestData(fixtureName, typeName)
		data, err := json.MarshalIndent(builder, "", "\t")
		Expect(err).ToNot(HaveOccurred())
		if record || (len(recordThisOne) > 0 && recordThisOne[0]) {
			writeTestFile(fixtureName, typeName, data)
		} else {
			loadData := loadTestFile(fixtureName, typeName)
			Expect(data).To(Equal(loadData))
		}
	},
		//Entry("Simple Struct", "testapp0_simple", "SimpleStruct"),
		//Entry("Simple Struct With Pointer Fields", "testapp0_simple", "SimpleStructWithPointer"),
		Entry("Struct with embedded field", "testapp0_simple", "StructWithEmbeddedField", true),
	)
})
