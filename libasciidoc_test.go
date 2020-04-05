package libasciidoc_test

import (
	"os"
	"time"

	"github.com/bytesparadise/libasciidoc/pkg/configuration"
	"github.com/bytesparadise/libasciidoc/pkg/types"
	. "github.com/bytesparadise/libasciidoc/testsupport"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("documents", func() {

	var level log.Level

	BeforeEach(func() {
		// turn down the logger to `warn` to avoid the noise
		level = log.GetLevel()
		log.SetLevel(log.WarnLevel)
	})

	AfterEach(func() {
		// restore the logger level
		log.SetLevel(level)
	})

	Context("document Body", func() {

		lastUpdated := time.Now()

		It("empty document", func() {
			// main title alone is not rendered in the body
			source := ""
			expectedContent := ""
			Expect(RenderHTML(source)).To(Equal(expectedContent))
			Expect(RenderHTML5Title(source)).To(Equal(""))
		})

		It("document with no section", func() {
			// main title alone is not rendered in the body
			source := "= a document title"
			expectedTitle := "a document title"
			expectedContent := ""
			Expect(RenderHTML(source)).To(Equal(expectedContent))
			Expect(RenderHTML5Title(source)).To(Equal(expectedTitle))
		})

		It("section levels 0 and 1", func() {
			source := `= a document title

== Section A

a paragraph with *bold content*`
			expectedTitle := "a document title"
			expectedContent := `<div class="sect1">
<h2 id="_section_a">Section A</h2>
<div class="sectionbody">
<div class="paragraph">
<p>a paragraph with <strong>bold content</strong></p>
</div>
</div>
</div>`
			Expect(RenderHTML(source)).To(Equal(expectedContent))
			Expect(RenderHTML5Title(source)).To(Equal(expectedTitle))
		})

		It("section level 1 with a paragraph", func() {
			source := `== Section A

a paragraph with *bold content*`
			expectedContent := `<div class="sect1">
<h2 id="_section_a">Section A</h2>
<div class="sectionbody">
<div class="paragraph">
<p>a paragraph with <strong>bold content</strong></p>
</div>
</div>
</div>`
			Expect(RenderHTML(source)).To(Equal(expectedContent))
			Expect(RenderHTML5Title(source)).To(Equal(""))
		})

		It("section levels 0, 1 and 3", func() {
			source := `= a document title

== Section A

a paragraph with *bold content*

==== Section A.a.a

a paragraph`
			expectedTitle := "a document title"
			expectedContent := `<div class="sect1">
<h2 id="_section_a">Section A</h2>
<div class="sectionbody">
<div class="paragraph">
<p>a paragraph with <strong>bold content</strong></p>
</div>
<div class="sect3">
<h4 id="_section_a_a_a">Section A.a.a</h4>
<div class="paragraph">
<p>a paragraph</p>
</div>
</div>
</div>
</div>`
			Expect(RenderHTML(source)).To(Equal(expectedContent))
			Expect(RenderHTML5Title(source)).To(Equal(expectedTitle))
			Expect(DocumentMetadata(source, lastUpdated)).To(Equal(types.Metadata{
				Title:       "a document title",
				LastUpdated: lastUpdated.Format(configuration.LastUpdatedFormat),
				TableOfContents: types.TableOfContents{
					Sections: []types.ToCSection{
						{
							ID:    "_section_a",
							Level: 1,
							Title: "Section A",
							Children: []types.ToCSection{
								{
									ID:       "_section_a_a_a",
									Level:    3,
									Title:    "Section A.a.a",
									Children: []types.ToCSection{},
								},
							},
						},
					},
				},
			}))
		})

		It("section levels 1, 2, 3 and 2", func() {
			source := `= a document title

== Section A

a paragraph with *bold content*

=== Section A.a

a paragraph

== Section B

a paragraph with _italic content_`
			expectedTitle := "a document title"
			expectedContent := `<div class="sect1">
<h2 id="_section_a">Section A</h2>
<div class="sectionbody">
<div class="paragraph">
<p>a paragraph with <strong>bold content</strong></p>
</div>
<div class="sect2">
<h3 id="_section_a_a">Section A.a</h3>
<div class="paragraph">
<p>a paragraph</p>
</div>
</div>
</div>
</div>
<div class="sect1">
<h2 id="_section_b">Section B</h2>
<div class="sectionbody">
<div class="paragraph">
<p>a paragraph with <em>italic content</em></p>
</div>
</div>
</div>`
			Expect(RenderHTML(source)).To(Equal(expectedContent))
			Expect(RenderHTML5Title(source)).To(Equal(expectedTitle))
			Expect(DocumentMetadata(source, lastUpdated)).To(Equal(types.Metadata{
				Title:       "a document title",
				LastUpdated: lastUpdated.Format(configuration.LastUpdatedFormat),
				TableOfContents: types.TableOfContents{
					Sections: []types.ToCSection{
						{
							ID:    "_section_a",
							Level: 1,
							Title: "Section A",
							Children: []types.ToCSection{
								{
									ID:       "_section_a_a",
									Level:    2,
									Title:    "Section A.a",
									Children: []types.ToCSection{},
								},
							},
						},
						{
							ID:       "_section_b",
							Level:    1,
							Title:    "Section B",
							Children: []types.ToCSection{},
						},
					},
				},
			}))
		})

		It("should include adoc file without leveloffset from local file", func() {
			source := "include::test/includes/grandchild-include.adoc[]"
			expected := `<div class="sect1">
<h2 id="_grandchild_title">grandchild title</h2>
<div class="sectionbody">
<div class="paragraph">
<p>first line of grandchild</p>
</div>
<div class="paragraph">
<p>last line of grandchild</p>
</div>
</div>
</div>`
			Expect(RenderHTML(source, configuration.WithFilename("test.adoc"), configuration.WithLastUpdated(lastUpdated))).To(Equal(expected))
			Expect(DocumentMetadata(source, lastUpdated)).To(Equal(types.Metadata{
				Title:       "",
				LastUpdated: lastUpdated.Format(configuration.LastUpdatedFormat),
				TableOfContents: types.TableOfContents{
					Sections: []types.ToCSection{
						{
							ID:       "_grandchild_title",
							Level:    1,
							Title:    "grandchild title",
							Children: []types.ToCSection{},
						},
					},
				},
			}))
		})
	})

	Context("complete Document ", func() {

		It("using existing file", func() {
			expectedContent := `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<!--[if IE]><meta http-equiv="X-UA-Compatible" content="IE=edge"><![endif]-->
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta name="generator" content="libasciidoc">
<link type="text/css" rel="stylesheet" href="path/to/style.css">
<title>Chapter A</title>
</head>
<body class="article">
<div id="header">
<h1>Chapter A</h1>
</div>
<div id="content">
<div class="paragraph">
<p>content</p>
</div>
</div>
<div id="footer">
<div id="footer-text">
Last updated {{.LastUpdated}}
</div>
</div>
</body>
</html>`
			filename := "test/includes/chapter-a.adoc"
			stat, err := os.Stat(filename)
			Expect(err).NotTo(HaveOccurred())
			Expect(RenderHTML5Document(filename, configuration.WithCSS("path/to/style.css"), configuration.WithHeaderFooter(true))).To(MatchHTML5Template(expectedContent, stat.ModTime()))
		})

	})

})
