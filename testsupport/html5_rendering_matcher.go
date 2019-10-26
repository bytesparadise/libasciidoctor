package testsupport

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bytesparadise/libasciidoc"
	"github.com/bytesparadise/libasciidoc/pkg/parser"
	"github.com/bytesparadise/libasciidoc/pkg/renderer"
	"github.com/bytesparadise/libasciidoc/pkg/renderer/html5"
	"github.com/bytesparadise/libasciidoc/pkg/types"

	gomegatypes "github.com/onsi/gomega/types"
	"github.com/pkg/errors"
)

// --------------------
// Render HTML5 Element
// --------------------

// RenderHTML5Element a custom matcher to verify that a block renders as the expectation
func RenderHTML5Element(expected string, options ...interface{}) gomegatypes.GomegaMatcher {
	m := &html5ElementMatcher{
		expected: expected,
		filename: "test.adoc",
		opts:     []renderer.Option{},
	}
	for _, o := range options {
		if configure, ok := o.(FilenameOption); ok {
			configure(m)
		} else if opt, ok := o.(renderer.Option); ok {
			m.opts = append(m.opts, opt)
		}
	}
	return m
}

func (m *html5ElementMatcher) setFilename(f string) {
	m.filename = f
}

type html5ElementMatcher struct {
	opts       []renderer.Option
	filename   string
	expected   string
	actual     string
	comparison comparison
}

func (m *html5ElementMatcher) Match(actual interface{}) (success bool, err error) {
	content, ok := actual.(string)
	if !ok {
		return false, errors.Errorf("RenderHTML5Element matcher expects a string (actual: %T)", actual)
	}
	r := strings.NewReader(content)
	doc, err := parser.ParseDocument(m.filename, r)
	if err != nil {
		return false, err
	}
	buff := bytes.NewBuffer(nil)
	rendererCtx := renderer.Wrap(context.Background(), doc, m.opts...)
	// insert tables of contents, preamble and process file inclusions
	err = renderer.Prerender(rendererCtx)
	if err != nil {
		return false, err
	}
	_, err = html5.Render(rendererCtx, buff)
	if err != nil {
		return false, err
	}
	if strings.Contains(m.expected, "{{.LastUpdated}}") {
		m.expected = strings.Replace(m.expected, "{{.LastUpdated}}", rendererCtx.LastUpdated(), 1)
	}
	m.actual = buff.String()
	m.comparison = compare(m.actual, m.expected)
	return m.comparison.diffs == "", nil
}

func (m *html5ElementMatcher) FailureMessage(_ interface{}) (message string) {
	return fmt.Sprintf("expected HTML5 elements to match:\n%s", m.comparison.diffs)
}

func (m *html5ElementMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	return fmt.Sprintf("expected HTML5 elements not to match:\n%s", m.comparison.diffs)
}

// --------------------
// Render HTML5 Body
// --------------------

// RenderHTML5Body a custom matcher to verify that a block renders as the expectation
func RenderHTML5Body(expected string, options ...interface{}) gomegatypes.GomegaMatcher {
	m := &html5BodyMatcher{
		expected: expected,
		filename: "test.adoc",
	}
	for _, o := range options {
		if configure, ok := o.(FilenameOption); ok {
			configure(m)
		}
	}
	return m
}

func (m *html5BodyMatcher) setFilename(f string) {
	m.filename = f
}

type html5BodyMatcher struct {
	filename string
	expected string
	actual   string
	opts     []renderer.Option
}

func (m *html5BodyMatcher) Match(actual interface{}) (success bool, err error) {
	content, ok := actual.(string)
	if !ok {
		return false, errors.Errorf("RenderHTML5Body matcher expects a string (actual: %T)", actual)
	}
	contentReader := strings.NewReader(content)
	resultWriter := bytes.NewBuffer(nil)
	_, err = libasciidoc.ConvertToHTML(context.Background(), m.filename, contentReader, resultWriter, renderer.IncludeHeaderFooter(false))
	if err != nil {
		return false, err
	}
	m.actual = resultWriter.String()
	return m.expected == m.actual, nil
}

func (m *html5BodyMatcher) FailureMessage(_ interface{}) (message string) {
	return fmt.Sprintf("expected HTML5 bodies to match:\n\texpected: '%v'\n\tactual:   '%v'", m.expected, m.actual)
}

func (m *html5BodyMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	return fmt.Sprintf("expected HTML5 bodies not to match:\n\texpected: '%v'\n\tactual:   '%v'", m.expected, m.actual)
}

// --------------------
// Render HTML5 Title
// --------------------

// RenderHTML5Title a custom matcher to verify that a block renders as the expectation
func RenderHTML5Title(expected interface{}, options ...interface{}) gomegatypes.GomegaMatcher {
	m := &html5TitleMatcher{
		expected: expected,
		filename: "test.adoc",
	}
	for _, o := range options {
		if configure, ok := o.(FilenameOption); ok {
			configure(m)
		}
	}
	return m
}

func (m *html5TitleMatcher) setFilename(f string) {
	m.filename = f
}

type html5TitleMatcher struct {
	filename string
	expected interface{}
	actual   interface{}
}

func (m *html5TitleMatcher) Match(actual interface{}) (success bool, err error) {
	content, ok := actual.(string)
	if !ok {
		return false, errors.Errorf("RenderHTML5Title matcher expects a string (actual: %T)", actual)
	}
	contentReader := strings.NewReader(content)
	resultWriter := bytes.NewBuffer(nil)
	metadata, err := libasciidoc.ConvertToHTML(context.Background(), m.filename, contentReader, resultWriter, renderer.IncludeHeaderFooter(false))
	if err != nil {
		return false, err
	}
	if metadata == nil {
		return false, errors.New("no metadata returned")
	}
	if m.expected == nil {
		actualTitle, found := metadata[types.AttrTitle]
		m.actual = actualTitle
		return !found, nil
	}

	actualTitle, ok := metadata[types.AttrTitle].(string)
	if !ok {
		return false, errors.Errorf("invalid type of title (%T)", metadata[types.AttrTitle])
	}
	m.actual = actualTitle
	return m.expected == m.actual, nil
}

func (m *html5TitleMatcher) FailureMessage(_ interface{}) (message string) {
	return fmt.Sprintf("expected HTML5 titles to match:\n\texpected: '%v'\n\tactual:   '%v'", m.expected, m.actual)
}

func (m *html5TitleMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	return fmt.Sprintf("expected HTML5 titles not to match:\n\texpected: '%v'\n\tactual:   '%v'", m.expected, m.actual)
}

// ---------------------
// Render HTML5 Document
// ---------------------

// RenderHTML5Document a custom matcher to verify that a block renders as the expectation
func RenderHTML5Document(expected string, options ...interface{}) gomegatypes.GomegaMatcher {
	m := &html5DocumentMatcher{
		expected: expected,
		filename: "test.adoc",
	}
	for _, o := range options {
		if configure, ok := o.(FilenameOption); ok {
			configure(m)
		}
	}
	return m
}

func (m *html5DocumentMatcher) setFilename(f string) {
	m.filename = f
}

type html5DocumentMatcher struct {
	filename string
	expected string
	actual   string
}

func (m *html5DocumentMatcher) Match(actual interface{}) (success bool, err error) {
	content, ok := actual.(string)
	if !ok {
		return false, errors.Errorf("RenderHTML5Body matcher expects a string (actual: %T)", actual)
	}
	contentReader := strings.NewReader(content)
	resultWriter := bytes.NewBuffer(nil)
	lastUpdated := time.Now()
	_, err = libasciidoc.ConvertToHTML(context.Background(), m.filename, contentReader, resultWriter, renderer.IncludeHeaderFooter(true))
	if err != nil {
		return false, err
	}
	m.expected = strings.Replace(m.expected, "{{.LastUpdated}}", lastUpdated.Format(renderer.LastUpdatedFormat), 1)
	m.actual = resultWriter.String()
	return m.expected == m.actual, nil
}

func (m *html5DocumentMatcher) FailureMessage(_ interface{}) (message string) {
	return fmt.Sprintf("expected HTML5 bodies to match:\n\texpected: '%v'\n\tactual:   '%v'", m.expected, m.actual)
}

func (m *html5DocumentMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	return fmt.Sprintf("expected HTML5 bodies not to match:\n\texpected: '%v'\n\tactual:   '%v'", m.expected, m.actual)
}
