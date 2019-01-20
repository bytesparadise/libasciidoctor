package html5_test

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("labeled lists of items", func() {

	Context("simple items", func() {

		It("simple labeled list with id, title, role and a default layout", func() {
			actualContent := `.mytitle
[#listID]
[.myrole]
item 1:: description 1.
item 2:: description 2
on 2 lines.
item 3:: description 3
on 2 lines, too.`
			expectedResult := `<div id="listID" class="dlist myrole">
<div class="title">mytitle</div>
<dl>
<dt class="hdlist1">item 1</dt>
<dd>
<p>description 1.</p>
</dd>
<dt class="hdlist1">item 2</dt>
<dd>
<p>description 2
on 2 lines.</p>
</dd>
<dt class="hdlist1">item 3</dt>
<dd>
<p>description 3
on 2 lines, too.</p>
</dd>
</dl>
</div>`
			verify(GinkgoT(), expectedResult, actualContent)
		})

		It("labeled list with an empty entry", func() {
			actualContent := `item 1::
item 2:: description 2.`
			expectedResult := `<div class="dlist">
<dl>
<dt class="hdlist1">item 1</dt>
<dt class="hdlist1">item 2</dt>
<dd>
<p>description 2.</p>
</dd>
</dl>
</div>`
			verify(GinkgoT(), expectedResult, actualContent)
		})

		It("labeled list with an image", func() {
			actualContent := `item 1:: image:foo.png[]
item 2:: description 2.`
			expectedResult := `<div class="dlist">
<dl>
<dt class="hdlist1">item 1</dt>
<dd>
<p><span class="image"><img src="foo.png" alt="foo"></span></p>
</dd>
<dt class="hdlist1">item 2</dt>
<dd>
<p>description 2.</p>
</dd>
</dl>
</div>`
			verify(GinkgoT(), expectedResult, actualContent)
		})

		It("labeled list with script injection", func() {
			actualContent := `item 1:: <script>alert("foo!")</script>`
			expectedResult := `<div class="dlist">
<dl>
<dt class="hdlist1">item 1</dt>
<dd>
<p>&lt;script&gt;alert(&#34;foo!&#34;)&lt;/script&gt;</p>
</dd>
</dl>
</div>`
			verify(GinkgoT(), expectedResult, actualContent)
		})

		It("labeled list with fenced block", func() {
			actualContent := "item 1::\n" +
				"```\n" +
				"a fenced block\n" +
				"```\n" +
				"item 2:: something simple"
			expectedResult := `<div class="dlist">
<dl>
<dt class="hdlist1">item 1</dt>
</dl>
</div>
<div class="listingblock">
<div class="content">
<pre class="highlight"><code>a fenced block</code></pre>
</div>
</div>
<div class="dlist">
<dl>
<dt class="hdlist1">item 2</dt>
<dd>
<p>something simple</p>
</dd>
</dl>
</div>`
			verify(GinkgoT(), expectedResult, actualContent)
		})

		It("labeled list with nested lists using regular layout", func() {
			actualContent := `item 1:: 
* foo
* bar
** baz
item 2:: something simple`
			expectedResult := `<div class="dlist">
<dl>
<dt class="hdlist1">item 1</dt>
<dd>
<div class="ulist">
<ul>
<li>
<p>foo</p>
</li>
<li>
<p>bar</p>
<div class="ulist">
<ul>
<li>
<p>baz</p>
</li>
</ul>
</div>
</li>
</ul>
</div>
</dd>
<dt class="hdlist1">item 2</dt>
<dd>
<p>something simple</p>
</dd>
</dl>
</div>`
			verify(GinkgoT(), expectedResult, actualContent)
		})

		It("labeled list with title", func() {
			actualContent := `.Labeled, single-line
first term:: definition of the first term
second term:: definition of the second term`
			expectedResult := `<div class="dlist">
<div class="title">Labeled, single-line</div>
<dl>
<dt class="hdlist1">first term</dt>
<dd>
<p>definition of the first term</p>
</dd>
<dt class="hdlist1">second term</dt>
<dd>
<p>definition of the second term</p>
</dd>
</dl>
</div>`
			verify(GinkgoT(), expectedResult, actualContent)
		})

	})

	Context("horizontal layout", func() {

		It("simple labeled list with horizontal layout, id, title and role", func() {
			actualContent := `.title
[#myid]
[.myrole]
[horizontal]
item 1::
item 2:: description 2 on 1 line.
item 3:: description 3
on 2 lines, too.`
			expectedResult := `<div id="myid" class="hdlist myrole">
<div class="title">title</div>
<table>
<tr>
<td class="hdlist1">
item 1
<br>
item 2
</td>
<td class="hdlist2">
<p>description 2 on 1 line.</p>
</td>
</tr>
<tr>
<td class="hdlist1">
item 3
</td>
<td class="hdlist2">
<p>description 3
on 2 lines, too.</p>
</td>
</tr>
</table>
</div>`
			verify(GinkgoT(), expectedResult, actualContent)
		})

		It("labeled list with nested lists using horizontal layout", func() {
			actualContent := `[horizontal]
item 1:: 
* foo
* bar
** baz
item 2:: something simple`
			expectedResult := `<div class="hdlist">
<table>
<tr>
<td class="hdlist1">
item 1
</td>
<td class="hdlist2">
<div class="ulist">
<ul>
<li>
<p>foo</p>
</li>
<li>
<p>bar</p>
<div class="ulist">
<ul>
<li>
<p>baz</p>
</li>
</ul>
</div>
</li>
</ul>
</div>
</td>
</tr>
<tr>
<td class="hdlist1">
item 2
</td>
<td class="hdlist2">
<p>something simple</p>
</td>
</tr>
</table>
</div>`
			verify(GinkgoT(), expectedResult, actualContent)
		})

	})

	Context("list item continuation", func() {
		It("labeled list with continuation", func() {
			actualContent := `Item 1::
+
----
a delimited block
----
Item 2:: something simple
+
----
another delimited block
----`
			expectedDocument := `<div class="dlist">
<dl>
<dt class="hdlist1">Item 1</dt>
<dd>
<div class="listingblock">
<div class="content">
<pre>a delimited block</pre>
</div>
</div>
</dd>
<dt class="hdlist1">Item 2</dt>
<dd>
<p>something simple</p>
<div class="listingblock">
<div class="content">
<pre>another delimited block</pre>
</div>
</div>
</dd>
</dl>
</div>`

			verify(GinkgoT(), expectedDocument, actualContent)
		})

		It("labeled list without continuation", func() {
			actualContent := `Item 1::
----
a delimited block
----
Item 2:: something simple
----
another delimited block
----`
			expectedDocument := `<div class="dlist">
<dl>
<dt class="hdlist1">Item 1</dt>
</dl>
</div>
<div class="listingblock">
<div class="content">
<pre>a delimited block</pre>
</div>
</div>
<div class="dlist">
<dl>
<dt class="hdlist1">Item 2</dt>
<dd>
<p>something simple</p>
</dd>
</dl>
</div>
<div class="listingblock">
<div class="content">
<pre>another delimited block</pre>
</div>
</div>`

			verify(GinkgoT(), expectedDocument, actualContent)
		})
	})

	Context("nestedt labelled list items", func() {

		It("labeled list with multiple nested items", func() {
			actualContent := `Item 1::
Item 1 description
Item 2:::
Item 2 description
Item 3::::
Item 3 description`
			expectedDocument := `<div class="dlist">
<dl>
<dt class="hdlist1">Item 1</dt>
<dd>
<p>Item 1 description</p>
<div class="dlist">
<dl>
<dt class="hdlist1">Item 2</dt>
<dd>
<p>Item 2 description</p>
<div class="dlist">
<dl>
<dt class="hdlist1">Item 3</dt>
<dd>
<p>Item 3 description</p>
</dd>
</dl>
</div>
</dd>
</dl>
</div>
</dd>
</dl>
</div>`
			verify(GinkgoT(), expectedDocument, actualContent)
		})
	})

	Context("q and a", func() {

		It("q and a with title", func() {
			actualContent := `.Q&A
[qanda]
What is libsciidoc?::
	An implementation of the AsciiDoc processor in Golang.
What is the answer to the Ultimate Question?:: 42`

			expectedDocument := `<div class="qlist qanda">
<div class="title">Q&amp;A</div>
<ol>
<li>
<p><em>What is libsciidoc?</em></p>
<p>An implementation of the AsciiDoc processor in Golang.</p>
</li>
<li>
<p><em>What is the answer to the Ultimate Question?</em></p>
<p>42</p>
</li>
</ol>
</div>`
			verify(GinkgoT(), expectedDocument, actualContent)
		})
	})
})
