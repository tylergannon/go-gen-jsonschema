package testapp2_test

import (
	"embed"
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	testapp2 "github.com/tylergannon/go-gen-jsonschema-testapp"
	"time"
)

type TestFunc func(character testapp2.MovieCharacter)

//go:embed fixtures
var fixtures embed.FS

var _ = DescribeTable("Unmarshaler test", func(fileName string, fn TestFunc) {
	data, err := fixtures.ReadFile(fmt.Sprintf("fixtures/%s.json", fileName))
	Expect(err).ToNot(HaveOccurred())

	var item testapp2.MovieCharacter
	err = json.Unmarshal(data, &item)
	Expect(err).NotTo(HaveOccurred())
	if fn != nil {
		fn(item)
	}
},
	Entry("fixture1", "fixture1", func(character testapp2.MovieCharacter) {
		Expect(character.Location).To(Equal(testapp2.HallyuWood))
		dob := time.Time(character.DateOfBirth)
		Expect(dob).NotTo(Equal(time.Time{}))
	}),
	Entry("fixture2", "fixture2", nil),
	Entry("fixture3", "fixture3", nil),
)
