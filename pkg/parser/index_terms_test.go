package parser_test

import (
	"github.com/bytesparadise/libasciidoc/pkg/types"
	. "github.com/bytesparadise/libasciidoc/testsupport"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("index terms", func() {

	Context("draft document", func() {

		It("index term in existing paragraph line", func() {
			source := `a paragraph with an ((index term)).`
			expected := types.DraftDocument{
				Blocks: []interface{}{
					types.Paragraph{
						Attributes: types.ElementAttributes{},
						Lines: [][]interface{}{
							{
								types.StringElement{
									Content: "a paragraph with an ",
								},
								types.IndexTerm{
									Term: []interface{}{
										types.StringElement{
											Content: "index term",
										},
									},
								},
								types.StringElement{
									Content: ".",
								},
							},
						},
					},
				},
			}
			Expect(ParseDraftDocument(source)).To(Equal(expected))
		})

		It("index term in single paragraph line", func() {
			source := `((_italic term_))
a paragraph with an index term.`
			expected := types.DraftDocument{
				Blocks: []interface{}{
					types.Paragraph{
						Attributes: types.ElementAttributes{},
						Lines: [][]interface{}{
							{
								types.IndexTerm{
									Term: []interface{}{
										types.QuotedText{
											Kind: types.Italic,
											Elements: []interface{}{
												types.StringElement{
													Content: "italic term",
												},
											},
										},
									},
								},
							},
							{
								types.StringElement{
									Content: "a paragraph with an index term.",
								},
							},
						},
					},
				},
			}
			Expect(ParseDraftDocument(source)).To(Equal(expected))
		})
	})

	Context("final document", func() {

		It("index term in existing paragraph line", func() {
			source := `a paragraph with an ((index)) term.`
			expected := types.Document{
				Attributes:         types.DocumentAttributes{},
				ElementReferences:  types.ElementReferences{},
				Footnotes:          types.Footnotes{},
				FootnoteReferences: types.FootnoteReferences{},
				Elements: []interface{}{
					types.Paragraph{
						Attributes: types.ElementAttributes{},
						Lines: [][]interface{}{
							{
								types.StringElement{
									Content: "a paragraph with an ",
								},
								types.IndexTerm{
									Term: []interface{}{types.StringElement{
										Content: "index",
									},
									},
								},
								types.StringElement{
									Content: " term.",
								},
							},
						},
					},
				},
			}
			Expect(ParseDocument(source)).To(Equal(expected))
		})

		It("index term in single paragraph line", func() {
			source := `((_italic_))
a paragraph with an index term.`
			expected := types.Document{
				Attributes:         types.DocumentAttributes{},
				ElementReferences:  types.ElementReferences{},
				Footnotes:          types.Footnotes{},
				FootnoteReferences: types.FootnoteReferences{},
				Elements: []interface{}{
					types.Paragraph{
						Attributes: types.ElementAttributes{},
						Lines: [][]interface{}{
							{
								types.IndexTerm{
									Term: []interface{}{
										types.QuotedText{
											Kind: types.Italic,
											Elements: []interface{}{
												types.StringElement{
													Content: "italic",
												},
											},
										},
									},
								},
							},
							{
								types.StringElement{
									Content: "a paragraph with an index term.",
								},
							},
						},
					},
				},
			}
			Expect(ParseDocument(source)).To(Equal(expected))
		})
	})
})
var _ = Describe("concealed index terms", func() {

	Context("draft document", func() {

		It("concealed index term in existing paragraph line", func() {
			source := `a paragraph with an index term (((index, term, here))).`
			expected := types.DraftDocument{
				Blocks: []interface{}{
					types.Paragraph{
						Attributes: types.ElementAttributes{},
						Lines: [][]interface{}{
							{
								types.StringElement{
									Content: "a paragraph with an index term ",
								},
								types.ConcealedIndexTerm{
									Term1: "index",
									Term2: "term",
									Term3: "here",
								},
								types.StringElement{
									Content: ".",
								},
							},
						},
					},
				},
			}
			Expect(ParseDraftDocument(source)).To(Equal(expected))
		})

		It("concealed index term in single paragraph line", func() {
			source := `(((index, term)))
a paragraph with an index term.`
			expected := types.DraftDocument{
				Blocks: []interface{}{
					types.Paragraph{
						Attributes: types.ElementAttributes{},
						Lines: [][]interface{}{
							{
								types.ConcealedIndexTerm{
									Term1: "index",
									Term2: "term",
								},
							},
							{
								types.StringElement{
									Content: "a paragraph with an index term.",
								},
							},
						},
					},
				},
			}
			Expect(ParseDraftDocument(source)).To(Equal(expected))
		})
	})

	Context("final document", func() {

		It("concealed index term in existing paragraph line", func() {
			source := `a paragraph with an index term (((index, term, here))).`
			expected := types.Document{
				Attributes:         types.DocumentAttributes{},
				ElementReferences:  types.ElementReferences{},
				Footnotes:          types.Footnotes{},
				FootnoteReferences: types.FootnoteReferences{},
				Elements: []interface{}{
					types.Paragraph{
						Attributes: types.ElementAttributes{},
						Lines: [][]interface{}{
							{
								types.StringElement{
									Content: "a paragraph with an index term ",
								},
								types.StringElement{
									Content: ".",
								},
							},
						},
					},
				},
			}
			Expect(ParseDocument(source)).To(Equal(expected))
		})

		It("concealed index term in single paragraph line", func() {
			source := `(((index, term)))
a paragraph with an index term.`
			expected := types.Document{
				Attributes:         types.DocumentAttributes{},
				ElementReferences:  types.ElementReferences{},
				Footnotes:          types.Footnotes{},
				FootnoteReferences: types.FootnoteReferences{},
				Elements: []interface{}{
					types.Paragraph{
						Attributes: types.ElementAttributes{},
						Lines: [][]interface{}{
							{
								types.StringElement{
									Content: "a paragraph with an index term.",
								},
							},
						},
					},
				},
			}
			Expect(ParseDocument(source)).To(Equal(expected))
		})
	})
})
