package testsupport_test

import (
	"fmt"

	"github.com/bytesparadise/libasciidoc/pkg/types"
	"github.com/bytesparadise/libasciidoc/testsupport"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("document assertions", func() {

	actual := "hello, world!"
	expected := types.Document{
		Attributes:         types.DocumentAttributes{},
		ElementReferences:  types.ElementReferences{},
		Footnotes:          types.Footnotes{},
		FootnoteReferences: types.FootnoteReferences{},
		Elements: []interface{}{
			types.Paragraph{
				Attributes: types.ElementAttributes{},
				Lines: []types.InlineElements{
					{
						types.StringElement{
							Content: "hello, world!",
						},
					},
				},
			},
		},
	}

	It("should match", func() {
		// given
		matcher := testsupport.EqualDocument(expected)
		// when
		result, err := matcher.Match(actual)
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeTrue())
	})

	It("should not match", func() {
		// given
		matcher := testsupport.EqualDocument(expected)
		// when
		result, err := matcher.Match("meh")
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeFalse())
	})

	Context("messages", func() {

		It("failure message", func() {
			// given
			matcher := testsupport.EqualDocument(expected)
			_, err := matcher.Match(actual)
			Expect(err).ToNot(HaveOccurred())
			// when
			msg := matcher.FailureMessage(actual)
			// then
			Expect(msg).To(Equal(fmt.Sprintf("expected documents to match:\n\texpected: '%v'\n\tactual:   '%v'", expected, expected)))
		})

		It("negated failure message", func() {
			// given
			matcher := testsupport.EqualDocument(expected)
			_, err := matcher.Match(actual)
			Expect(err).ToNot(HaveOccurred())
			// when
			msg := matcher.NegatedFailureMessage(actual)
			// then
			Expect(msg).To(Equal(fmt.Sprintf("expected documents not to match:\n\texpected: '%v'\n\tactual:   '%v'", expected, expected)))

		})
	})

	Context("failures", func() {

		It("should return error when invalid type is input", func() {
			// given
			matcher := testsupport.EqualDocument(types.Document{})
			_, err := matcher.Match(actual)
			Expect(err).ToNot(HaveOccurred())
			// when
			result, err := matcher.Match(1) // not a string
			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("EqualDocument matcher expects a string (actual: int)"))
			Expect(result).To(BeFalse())
		})
	})
})
