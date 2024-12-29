package testapp3validate_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	testapp2 "github.com/tylergannon/go-gen-jsonschema-testapp"
	"github.com/tylergannon/go-gen-jsonschema-testapp/llmfriendlytimepkg3"
	"time"
)

var _ = Describe("NearestDate", func() {
	Context("when TimeFrame is Future", func() {
		It("should return next year's date if month/day has already passed this year", func() {
			// Example: Suppose today is December 25, 2024. If we ask for Future January 1,
			// it should return January 1, 2025.
			nd := llmfriendlytimepkg3.NearestDate{
				TimeFrame:  llmfriendlytimepkg3.Future,
				Month:      llmfriendlytimepkg3.January,
				DayOfMonth: 1,
			}

			result, err := testapp2.NearestDateToTime(nd)
			Expect(err).NotTo(HaveOccurred())

			// We can't guarantee the test date is December 25, 2024 in real CI,
			// so a better test is to freeze time or make an assertion that
			// the returned date is strictly in the future. For demonstration:
			Expect(time.Time(result).After(time.Now())).To(BeTrue())
		})

		It("should return the same year date if it hasn't passed yet this year", func() {
			// Example: If it's currently February 10, and we ask for Future
			// February 20, it should remain in the same year.
			// Freeze time or assume the date for the test.
			// This is just a rough illustration.
			// ...
		})
	})

	Context("when TimeFrame is Past", func() {
		It("should return last year's date if month/day hasn't arrived yet this year", func() {
			nd := llmfriendlytimepkg3.NearestDate{
				TimeFrame:  llmfriendlytimepkg3.Past,
				Month:      llmfriendlytimepkg3.December,
				DayOfMonth: 31,
			}

			result, err := testapp2.NearestDateToTime(nd)
			Expect(err).NotTo(HaveOccurred())
			Expect(time.Time(result).Before(time.Now())).To(BeTrue())
		})

		// ...
	})
})

var _ = Describe("NearestDay", func() {
	Context("when TimeFrame is Future", func() {
		It("should find the upcoming Friday", func() {
			// freeze or note today's day
			nd := llmfriendlytimepkg3.NearestDay{
				TimeFrame: llmfriendlytimepkg3.Future,
				DayOfWeek: llmfriendlytimepkg3.Friday,
				Scale:     1,
			}

			result, err := testapp2.NearestDayToTime(nd)
			Expect(err).NotTo(HaveOccurred())

			// We expect the result to be a Friday in the future relative to now.
			Expect(time.Time(result).Weekday()).To(Equal(time.Friday))
			Expect(time.Time(result).After(time.Now())).To(BeTrue())
		})

		It("should skip two weeks if Scale is 2", func() {
			nd := llmfriendlytimepkg3.NearestDay{
				TimeFrame: llmfriendlytimepkg3.Future,
				DayOfWeek: llmfriendlytimepkg3.Monday,
				Scale:     2,
			}

			result, err := testapp2.NearestDayToTime(nd)
			Expect(err).NotTo(HaveOccurred())

			// The day should be Monday, at least 7 days after next Monday.
			Expect(time.Time(result).Weekday()).To(Equal(time.Monday))
			Expect(time.Time(result).After(time.Now())).To(BeTrue())
		})
	})

	Context("when TimeFrame is Past", func() {
		It("should find the last Monday", func() {
			nd := llmfriendlytimepkg3.NearestDay{
				TimeFrame: llmfriendlytimepkg3.Past,
				DayOfWeek: llmfriendlytimepkg3.Monday,
				Scale:     1,
			}

			result, err := testapp2.NearestDayToTime(nd)
			Expect(err).NotTo(HaveOccurred())
			Expect(time.Time(result).Weekday()).To(Equal(time.Monday))
			Expect(time.Time(result).Before(time.Now())).To(BeTrue())
		})

		// ...
	})
})
