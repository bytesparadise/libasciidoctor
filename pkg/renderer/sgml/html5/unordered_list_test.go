package html5_test

import (
	. "github.com/bytesparadise/libasciidoc/testsupport"

	. "github.com/onsi/ginkgo" //nolint golint
	. "github.com/onsi/gomega" //nolint golint
)

var _ = Describe("unordered lists", func() {

	It("simple unordered list with no title", func() {
		source := `* item 1
* item 2
* item 3`
		expected := `<div class="ulist">
<ul>
<li>
<p>item 1</p>
</li>
<li>
<p>item 2</p>
</li>
<li>
<p>item 3</p>
</li>
</ul>
</div>
`
		Expect(RenderHTML(source)).To(MatchHTML(expected))
	})

	It("simple unordered list with no title then a paragraph", func() {
		source := `* item 1
* item 2
* item 3

and a standalone paragraph`
		expected := `<div class="ulist">
<ul>
<li>
<p>item 1</p>
</li>
<li>
<p>item 2</p>
</li>
<li>
<p>item 3</p>
</li>
</ul>
</div>
<div class="paragraph">
<p>and a standalone paragraph</p>
</div>
`
		Expect(RenderHTML(source)).To(MatchHTML(expected))
	})

	It("simple unordered list with id, title and role", func() {
		source := `.mytitle
[#foo]
[.myrole]
* item 1
* item 2`
		expected := `<div id="foo" class="ulist myrole">
<div class="title">mytitle</div>
<ul>
<li>
<p>item 1</p>
</li>
<li>
<p>item 2</p>
</li>
</ul>
</div>
`
		Expect(RenderHTML(source)).To(MatchHTML(expected))
	})

	It("simple unordered list with id, title and role", func() {
		source := `.mytitle
[#foo]
[.myrole]
* item 1
* item 2`
		expected := `<div id="foo" class="ulist myrole">
<div class="title">mytitle</div>
<ul>
<li>
<p>item 1</p>
</li>
<li>
<p>item 2</p>
</li>
</ul>
</div>
`
		Expect(RenderHTML(source)).To(MatchHTML(expected))
	})

	It("simple unordered list with style id, title and role", func() {
		source := `.mytitle
[#foo]
[disc.myrole]
* item 1
* item 2`
		expected := `<div id="foo" class="ulist disc myrole">
<div class="title">mytitle</div>
<ul class="disc">
<li>
<p>item 1</p>
</li>
<li>
<p>item 2</p>
</li>
</ul>
</div>
`
		Expect(RenderHTML(source)).To(MatchHTML(expected))
	})

	It("unordered list item with dash on multiple lines", func() {
		source := `- an item (quite
  short) breaks` // with leading spaces which shall be trimmed during rendering
		expected := `<div class="ulist">
<ul>
<li>
<p>an item (quite
short) breaks</p>
</li>
</ul>
</div>
`
		Expect(RenderHTML(source)).To(MatchHTML(expected))
	})

	It("simple unordered list with continuation", func() {
		source := `* item 1
+
foo

* item 2`
		expected := `<div class="ulist">
<ul>
<li>
<p>item 1</p>
<div class="paragraph">
<p>foo</p>
</div>
</li>
<li>
<p>item 2</p>
</li>
</ul>
</div>
`
		Expect(RenderHTML(source)).To(MatchHTML(expected))
	})

	It("nested unordered lists without a title", func() {
		source := `* item 1
** item 1.1
** item 1.2
* item 2`
		expected := `<div class="ulist">
<ul>
<li>
<p>item 1</p>
<div class="ulist">
<ul>
<li>
<p>item 1.1</p>
</li>
<li>
<p>item 1.2</p>
</li>
</ul>
</div>
</li>
<li>
<p>item 2</p>
</li>
</ul>
</div>
`
		Expect(RenderHTML(source)).To(MatchHTML(expected))
	})

	It("nested unordered lists with a title", func() {
		source := `[#listID]
* item 1
** item 1.1
** item 1.2
* item 2`
		expected := `<div id="listID" class="ulist">
<ul>
<li>
<p>item 1</p>
<div class="ulist">
<ul>
<li>
<p>item 1.1</p>
</li>
<li>
<p>item 1.2</p>
</li>
</ul>
</div>
</li>
<li>
<p>item 2</p>
</li>
</ul>
</div>
`
		Expect(RenderHTML(source)).To(MatchHTML(expected))
	})

	It("unordered list with item continuation", func() {
		source := `* foo
+
----
a delimited block
----
+
----
another delimited block
----
* bar
`
		expected := `<div class="ulist">
<ul>
<li>
<p>foo</p>
<div class="listingblock">
<div class="content">
<pre>a delimited block</pre>
</div>
</div>
<div class="listingblock">
<div class="content">
<pre>another delimited block</pre>
</div>
</div>
</li>
<li>
<p>bar</p>
</li>
</ul>
</div>
`
		Expect(RenderHTML(source)).To(MatchHTML(expected))
	})

	It("unordered list without item continuation", func() {
		source := `* foo
----
a delimited block
----
* bar
----
another delimited block
----`
		expected := `<div class="ulist">
<ul>
<li>
<p>foo</p>
</li>
</ul>
</div>
<div class="listingblock">
<div class="content">
<pre>a delimited block</pre>
</div>
</div>
<div class="ulist">
<ul>
<li>
<p>bar</p>
</li>
</ul>
</div>
<div class="listingblock">
<div class="content">
<pre>another delimited block</pre>
</div>
</div>
`
		Expect(RenderHTML(source)).To(MatchHTML(expected))
	})
})
