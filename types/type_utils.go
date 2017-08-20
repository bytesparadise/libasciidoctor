package types

import (
	"bytes"
	"reflect"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func indent(indentLevel int) string {
	return strings.Repeat("  ", indentLevel)
}

func toDocElements(elements []interface{}) ([]DocElement, error) {
	result := make([]DocElement, len(elements))
	for i, element := range elements {
		switch element := element.(type) {
		case DocElement:
			result[i] = element
		default:
			return nil, errors.Errorf("unexpected element type: %v (expected a DocElement instead)", reflect.TypeOf(element))
		}
	}
	return result, nil
}

// convertBlocksToDocElements converts the given blocks to DocElement and exclude `BlankLine`
func convertBlocksToDocElements(blocks []interface{}) []DocElement {
	elements := make([]DocElement, len(blocks))
	j := 0
	for i := range blocks {
		// exclude blank lines from here, we won't need them in the rendering anyways
		if _, ok := blocks[i].(*BlankLine); !ok {
			elements[j] = blocks[i].(DocElement)
			j++
		}
	}
	return elements[:j] // exclude allocated nil values
}

func merge(elements []interface{}, extraElements ...interface{}) []interface{} {
	result := make([]interface{}, 0)
	allElements := append(elements, extraElements...)
	// log.Debugf("Merging %d element(s):", len(allElements))
	buff := bytes.NewBuffer(make([]byte, 0))
	for _, v := range allElements {
		if v == nil {
			continue
		}
		switch v.(type) {
		case string:
			buff.WriteString(v.(string))
		case []byte:
			for _, b := range v.([]byte) {
				buff.WriteByte(b)
			}
		case StringElement:
			content := v.(StringElement).Content
			buff.WriteString(content)
		case *StringElement:
			content := v.(*StringElement).Content
			buff.WriteString(content)
		case []interface{}:
			w := v.([]interface{})
			if len(w) > 0 {
				f := merge(w)
				result, buff = appendBuffer(result, buff)
				result = merge(result, f...)
			}
		default:
			result, buff = appendBuffer(result, buff)
			result = append(result, v.(DocElement))
		}
	}
	// if buff was filled because some text was found
	result, buff = appendBuffer(result, buff)
	// if len(extraElements) > 0 {
	// 	log.Debugf("merged '%v' (len=%d) with '%v' (len=%d) -> '%v' (len=%d)", elements, len(elements), extraElements, len(extraElements), result, len(result))

	// } else {
	// 	log.Debugf("merged '%v' (len=%d) -> '%v' (len=%d)", elements, len(elements), result, len(result))
	// }
	return result
}

// appendBuffer appends the content of the given buffer to the given array of elements, and returns a new buffer, or returns
// the given arguments if the buffer was empty
func appendBuffer(elements []interface{}, buff *bytes.Buffer) ([]interface{}, *bytes.Buffer) {
	if buff.Len() > 0 {
		return append(elements, NewStringElement(buff.String())), bytes.NewBuffer(make([]byte, 0))
	}
	return elements, buff
}

func stringify(elements []interface{}) (*string, error) {
	mergedElements := merge(elements)
	b := make([]byte, 0)
	buff := bytes.NewBuffer(b)
	for _, element := range mergedElements {
		if element == nil {
			continue
		}
		log.Debugf("%v (%s) ", element, reflect.TypeOf(element))
		switch element := element.(type) {
		case string:
			buff.WriteString(element)
		case []byte:
			for _, b := range element {
				buff.WriteByte(b)
			}
		case StringElement:
			buff.WriteString(element.Content)
		case *StringElement:
			buff.WriteString(element.Content)
		case []interface{}:
			stringifiedElement, err := stringify(element)
			if err != nil {
				// no need to wrap the error again in the same function
				return nil, err
			}
			buff.WriteString(*stringifiedElement)
		default:
			return nil, errors.Errorf("cannot convert element of type '%v' to string content", reflect.TypeOf(element))
		}

	}
	result := buff.String()
	log.Debugf("stringified %v -> '%s' (%v characters)", elements, result, len(result))
	return &result, nil
}

//NormalizationFunc a function that is used to normalize a string.
type NormalizationFunc func(string) ([]byte, error)

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

//ReplaceNonAlphanumerics replaces all non alphanumerical characters and remove (accents)
// in the given 'source' with the given 'replacement'.
func NewReplaceNonAlphanumericsFunc(replacement string) NormalizationFunc {
	return func(source string) ([]byte, error) {
		buf := bytes.NewBuffer(make([]byte, 0))
		// _, err := buf.WriteString(replacement)
		// if err != nil {
		// 	return nil, errors.Wrapf(err, "unable to normalize value")
		// }
		lastCharIsSpace := false
		for _, r := range strings.TrimLeft(source, " ") { // ignore heading spaces
			if unicode.Is(unicode.Letter, r) || unicode.Is(unicode.Number, r) {
				_, err := buf.WriteString(strings.ToLower(string(r)))
				if err != nil {
					return nil, errors.Wrapf(err, "unable to normalize value")
				}
				lastCharIsSpace = false
			} else if !lastCharIsSpace && (unicode.Is(unicode.Space, r) || unicode.Is(unicode.Punct, r)) {
				_, err := buf.WriteString(replacement)
				if err != nil {
					return nil, errors.Wrapf(err, "unable to normalize value")
				}
				lastCharIsSpace = true
			}
		}
		result := strings.TrimSuffix(buf.String(), replacement)
		return []byte(result), nil
	}
}

func ReplaceNonAlphanumerics(source *InlineContent, replacement string) (*string, error) {
	v := NewReplaceNonAlphanumericsVisitor()
	s := *source
	err := s.Accept(v)
	if err != nil {
		return nil, err
	}
	result := v.NormalizedContent()
	return &result, nil
}

//ReplaceNonAlphanumericsVisitor a visitor that builds a string representation of the visited elements,
// in which all non-alphanumeric characters have been replaced with a "_"
type ReplaceNonAlphanumericsVisitor struct {
	buf       bytes.Buffer
	normalize NormalizationFunc
}

func NewReplaceNonAlphanumericsVisitor() *ReplaceNonAlphanumericsVisitor {
	buf := bytes.NewBuffer(make([]byte, 0))
	return &ReplaceNonAlphanumericsVisitor{
		buf:       *buf,
		normalize: NewReplaceNonAlphanumericsFunc("_"),
	}
}

func (v *ReplaceNonAlphanumericsVisitor) Visit(element interface{}) error {
	switch element := element.(type) {
	case *InlineContent:
		// log.Debugf("Prefixing with '_' while processing '%v'", reflect.TypeOf(element))
		v.buf.WriteString("_")
	case *StringElement:
		normalized, err := v.normalize(element.Content)
		if err != nil {
			return errors.Wrapf(err, "error while normalizing String Element")
		}
		v.buf.Write(normalized)
	default:
		// ignore
	}
	return nil
}

func (v *ReplaceNonAlphanumericsVisitor) BeforeVisit(element interface{}) error {
	// log.Debugf("Before visiting element of type '%v'...", reflect.TypeOf(element))
	switch element := element.(type) {
	case *QuotedText:
		// log.Debugf("Before visiting quoted element...")
		switch element.Kind {
		case Bold:
			v.buf.WriteString("_strong_")
		case Italic:
			v.buf.WriteString("_italic_")
		case Monospace:
			v.buf.WriteString("_monospace_")
		default:
			return errors.Errorf("unsupported kind of quoted text: %d", element.Kind)
		}
	default:
		// ignore
	}
	return nil
}

func (v *ReplaceNonAlphanumericsVisitor) AfterVisit(element interface{}) error {
	switch element := element.(type) {
	case *QuotedText:
		switch element.Kind {
		case Bold:
			v.buf.WriteString("_strong")
		case Italic:
			v.buf.WriteString("_italic")
		case Monospace:
			v.buf.WriteString("_monospace")
		default:
			return errors.Errorf("unsupported kind of quoted text: %d", element.Kind)
		}
	default:
		// ignore
	}
	return nil
}

func (v *ReplaceNonAlphanumericsVisitor) NormalizedContent() string {
	result := v.buf.String()
	return result
}
