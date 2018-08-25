package types

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ------------------------------------------
// interface{} (and other interfaces)
// ------------------------------------------

// Visitable the interface for visitable elements
type Visitable interface {
	Accept(Visitor) error
}

// Visitor a visitor that can visit/traverse the interface{} and its children (if applicable)
type Visitor interface {
	BeforeVisit(Visitable) error
	Visit(Visitable) error
	AfterVisit(Visitable) error
}

// ------------------------------------------
// Document
// ------------------------------------------

// Document the top-level structure for a document
type Document struct {
	Attributes        DocumentAttributes
	Elements          []interface{}
	ElementReferences ElementReferences
}

// NewDocument initializes a new `Document` from the given lines
func NewDocument(frontmatter, header interface{}, blocks []interface{}) (Document, error) {
	log.Debugf("initializing a new Document with %d blocks(s)", len(blocks))
	// elements := convertBlocksTointerface{}s(blocks)
	// elements := filterEmptyElements(blocks, filterBlankLine(), filterEmptyPreamble())
	elements := insertPreamble(blocks)
	attributes := make(DocumentAttributes)

	if frontmatter != nil {
		for attrName, attrValue := range frontmatter.(FrontMatter).Content {
			attributes[attrName] = attrValue
		}
	}
	if header != nil {
		for attrName, attrValue := range header.(DocumentHeader).Content {
			attributes[attrName] = attrValue
			if attrName == "toc" {
				// insert a TableOfContentsMacro element if `toc` value is:
				// - "auto" (or empty)
				// - "preamble"
				switch attrValue {
				case "", "auto":
					// insert TableOfContentsMacro at first position
					elements = append([]interface{}{TableOfContentsMacro{}}, elements...)
				case "preamble":
					// lookup preamble in elements (should be first)
					preambleIndex := 0
					for i, e := range elements {
						if _, ok := e.(Preamble); ok {
							preambleIndex = i
							break
						}
					}
					// insert TableOfContentsMacro just after preamble
					remainingElements := make([]interface{}, len(elements)-(preambleIndex+1))
					copy(remainingElements, elements[preambleIndex+1:])
					elements = append(elements[0:preambleIndex+1], TableOfContentsMacro{})
					elements = append(elements, remainingElements...)
				case "macro":
				default:
					log.Warnf("invalid value for 'toc' attribute: '%s'", attrValue)

				}
			}
		}
	}

	c := NewElementReferencesCollector()
	for _, e := range elements {
		if v, ok := e.(Visitable); ok {
			v.Accept(c)
		}
	}
	document := Document{
		Attributes:        attributes,
		Elements:          elements,
		ElementReferences: c.ElementReferences,
	}

	// visit all elements in the `AST` to retrieve their reference (ie, their ElementID if they have any)
	return document, nil
}

func insertPreamble(blocks []interface{}) []interface{} {
	// log.Debugf("generating preamble from %d blocks", len(blocks))
	preamble := NewEmptyPreamble()
	for _, block := range blocks {
		switch block.(type) {
		case Section:
			break
		default:
			preamble.Elements = append(preamble.Elements, block)
		}
	}
	// no element in the preamble, or no section in the document, so no preamble to generate
	if len(preamble.Elements) == 0 || len(preamble.Elements) == len(blocks) {
		log.Debugf("skipping preamble (%d vs %d)", len(preamble.Elements), len(blocks))
		return nilSafe(blocks)
	}
	// now, insert the preamble instead of the 'n' blocks that belong to the preamble
	// and copy the other items
	result := make([]interface{}, len(blocks)-len(preamble.Elements)+1)
	result[0] = preamble
	copy(result[1:], blocks[len(preamble.Elements):])
	log.Debugf("generated preamble with %d blocks", len(preamble.Elements))
	return result
}

// ------------------------------------------
// Document Header
// ------------------------------------------

// DocumentHeader the document header
type DocumentHeader struct {
	Content DocumentAttributes
}

// NewDocumentHeader initializes a new DocumentHeader
func NewDocumentHeader(header, authors, revision interface{}, otherAttributes []interface{}) (DocumentHeader, error) {
	content := DocumentAttributes{}
	if header != nil {
		content["doctitle"] = header.(SectionTitle)
	}
	log.Debugf("initializing a new DocumentHeader with content '%v', authors '%+v' and revision '%+v'", content, authors, revision)
	if authors != nil {
		for i, author := range authors.([]DocumentAuthor) {
			if i == 0 {
				content.AddNonEmpty("firstname", author.FirstName)
				content.AddNonEmpty("middlename", author.MiddleName)
				content.AddNonEmpty("lastname", author.LastName)
				content.AddNonEmpty("author", author.FullName)
				content.AddNonEmpty("authorinitials", author.Initials)
				content.AddNonEmpty("email", author.Email)
			} else {
				content.AddNonEmpty(fmt.Sprintf("firstname_%d", i+1), author.FirstName)
				content.AddNonEmpty(fmt.Sprintf("middlename_%d", i+1), author.MiddleName)
				content.AddNonEmpty(fmt.Sprintf("lastname_%d", i+1), author.LastName)
				content.AddNonEmpty(fmt.Sprintf("author_%d", i+1), author.FullName)
				content.AddNonEmpty(fmt.Sprintf("authorinitials_%d", i+1), author.Initials)
				content.AddNonEmpty(fmt.Sprintf("email_%d", i+1), author.Email)
			}
		}
	}
	if revision != nil {
		rev := revision.(DocumentRevision)
		content.AddNonEmpty("revnumber", rev.Revnumber)
		content.AddNonEmpty("revdate", rev.Revdate)
		content.AddNonEmpty("revremark", rev.Revremark)
	}
	for _, attr := range otherAttributes {
		if attr, ok := attr.(DocumentAttributeDeclaration); ok {
			content.AddAttribute(attr)
		}
	}
	return DocumentHeader{
		Content: content,
	}, nil
}

// ------------------------------------------
// Document Author
// ------------------------------------------

// DocumentAuthor a document author
type DocumentAuthor struct {
	FullName   string
	Initials   string
	FirstName  string
	MiddleName string
	LastName   string
	Email      string
}

// NewDocumentAuthors converts the given authors into an array of `DocumentAuthor`
func NewDocumentAuthors(authors []interface{}) ([]DocumentAuthor, error) {
	log.Debugf("initializing a new array of document authors from `%+v`", authors)
	result := make([]DocumentAuthor, len(authors))
	for i, author := range authors {
		switch author.(type) {
		case DocumentAuthor:
			result[i] = author.(DocumentAuthor)
		default:
			return nil, errors.Errorf("unexpected type of author: %T", author)
		}
	}
	return result, nil
}

//NewDocumentAuthor initializes a new DocumentAuthor
func NewDocumentAuthor(namePart1, namePart2, namePart3, emailAddress interface{}) (DocumentAuthor, error) {
	var part1, part2, part3, email string
	if namePart1, ok := namePart1.(string); ok {
		part1 = apply(namePart1,
			func(s string) string {
				return strings.TrimSpace(s)
			},
			func(s string) string {
				return strings.Replace(s, "_", " ", -1)
			},
		)
	}
	if namePart2, ok := namePart2.(string); ok {
		part2 = apply(namePart2,
			func(s string) string {
				return strings.TrimSpace(s)
			},
			func(s string) string {
				return strings.Replace(s, "_", " ", -1)
			},
		)
	}
	if namePart3, ok := namePart3.(string); ok {
		part3 = apply(namePart3,
			func(s string) string {
				return strings.TrimSpace(s)
			},
			func(s string) string {
				return strings.Replace(s, "_", " ", -1)
			},
		)
	}
	if emailAddress, ok := emailAddress.(string); ok {
		email = apply(emailAddress,
			func(s string) string {
				return strings.TrimPrefix(s, "<")
			}, func(s string) string {
				return strings.TrimSuffix(s, ">")
			}, func(s string) string {
				return strings.TrimSpace(s)
			})
	}
	result := DocumentAuthor{}
	if part2 != "" && part3 != "" {
		result.FirstName = part1
		result.MiddleName = part2
		result.LastName = part3
		result.FullName = fmt.Sprintf("%s %s %s", part1, part2, part3)
		result.Initials = initials(result.FirstName, result.MiddleName, result.LastName)
	} else if part2 != "" {
		result.FirstName = part1
		result.LastName = part2
		result.FullName = fmt.Sprintf("%s %s", part1, part2)
		result.Initials = initials(result.FirstName, result.LastName)
	} else {
		result.FirstName = part1
		result.FullName = part1
		result.Initials = initials(result.FirstName)
	}
	result.Email = email
	log.Debugf("Initialized a new document author: `%v`", result.String())
	return result, nil
}

func initials(firstPart string, otherParts ...string) string {
	result := firstPart[0:1]
	for _, otherPart := range otherParts {
		result = result + otherPart[0:1]
	}
	return result
}

func (a *DocumentAuthor) String() string {
	email := ""
	if a.Email != "" {
		email = a.Email
	}
	return fmt.Sprintf("%s (%s)", a.FullName, email)
}

// ------------------------------------------
// Document Revision
// ------------------------------------------

// DocumentRevision a document revision
type DocumentRevision struct {
	Revnumber string
	Revdate   string
	Revremark string
}

// NewDocumentRevision intializes a new DocumentRevision
func NewDocumentRevision(revnumber, revdate, revremark interface{}) (DocumentRevision, error) {
	log.Debugf("initializing document revision with revnumber=%v, revdate=%v, revremark=%v", revnumber, revdate, revremark)
	// remove the "v" prefix and trim spaces
	var number, date, remark string
	if revnumber, ok := revnumber.(string); ok {
		number = apply(revnumber,
			func(s string) string {
				return strings.TrimPrefix(s, "v")
			}, func(s string) string {
				return strings.TrimPrefix(s, "V")
			}, func(s string) string {
				return strings.TrimSpace(s)
			})
	}
	if revdate, ok := revdate.(string); ok {
		// trim spaces
		date = apply(revdate,
			func(s string) string {
				return strings.TrimSpace(s)
			})
	}
	if revremark, ok := revremark.(string); ok {
		// then we need to strip the heading ":" and spaces
		remark = apply(revremark,
			func(s string) string {
				return strings.TrimPrefix(s, ":")
			}, func(s string) string {
				return strings.TrimSpace(s)
			})
	}
	// log.Debugf("initializing a new DocumentRevision with revnumber='%v', revdate='%v' and revremark='%v'", *n, *d, *r)
	result := DocumentRevision{
		Revnumber: number,
		Revdate:   date,
		Revremark: remark,
	}
	log.Debugf("Initialized a new document revision: `%s`", result.String())
	return result, nil
}

func (r DocumentRevision) String() string {
	// return fmt.Sprintf("%v, %v: %v", number, date, remark)
	return fmt.Sprintf("%v, %v: %v", r.Revnumber, r.Revdate, r.Revremark)
}

// ------------------------------------------
// Document Attributes
// ------------------------------------------

// DocumentAttributeDeclaration the type for Document Attribute Declarations
type DocumentAttributeDeclaration struct {
	Name  string
	Value string
}

// NewDocumentAttributeDeclaration initializes a new DocumentAttributeDeclaration with the given name and optional value
func NewDocumentAttributeDeclaration(name string, value interface{}) (DocumentAttributeDeclaration, error) {
	var attrName, attrValue string
	attrName = apply(name,
		func(s string) string {
			return strings.TrimSpace(s)
		})
	if value, ok := value.(string); ok {
		attrValue = apply(value,
			func(s string) string {
				return strings.TrimSpace(s)
			})
	}
	log.Debugf("Initialized a new DocumentAttributeDeclaration: '%s' -> '%s'", attrName, attrValue)
	return DocumentAttributeDeclaration{
		Name:  attrName,
		Value: attrValue,
	}, nil
}

// DocumentAttributeReset the type for DocumentAttributeReset
type DocumentAttributeReset struct {
	Name string
}

// NewDocumentAttributeReset initializes a new Document Attribute Resets.
func NewDocumentAttributeReset(attrName string) (DocumentAttributeReset, error) {
	log.Debugf("Initialized a new DocumentAttributeReset: '%s'", attrName)
	return DocumentAttributeReset{Name: attrName}, nil
}

// DocumentAttributeSubstitution the type for DocumentAttributeSubstitution
type DocumentAttributeSubstitution struct {
	Name string
}

// NewDocumentAttributeSubstitution initializes a new Document Attribute Substitutions
func NewDocumentAttributeSubstitution(attrName string) (DocumentAttributeSubstitution, error) {
	log.Debugf("Initialized a new DocumentAttributeSubstitution: '%s'", attrName)
	return DocumentAttributeSubstitution{Name: attrName}, nil
}

// ------------------------------------------
// Element kinds
// ------------------------------------------

// BlockKind the kind of block
type BlockKind int

const (
	// AttrKind the key for the kind of block
	AttrKind string = "kind"
	// Fenced a fenced block
	Fenced BlockKind = iota // 1
	// Listing a listing block
	Listing
	// Example an example block
	Example
	// Comment a comment block
	Comment
	// Quote a quote block
	Quote
	// Verse a verse block
	Verse
)

// ------------------------------------------
// Table of Contents
// ------------------------------------------

// TableOfContentsMacro the structure for Table of Contents
type TableOfContentsMacro struct {
}

// ------------------------------------------
// Preamble
// ------------------------------------------

// Preamble the structure for document Preamble
type Preamble struct {
	Elements []interface{}
}

// NewEmptyPreamble return an empty Preamble
func NewEmptyPreamble() Preamble {
	return Preamble{
		Elements: make([]interface{}, 0),
	}
}

// ------------------------------------------
// Front Matter
// ------------------------------------------

// FrontMatter the structure for document front-matter
type FrontMatter struct {
	Content map[string]interface{}
}

// NewYamlFrontMatter initializes a new FrontMatter from the given `content`
func NewYamlFrontMatter(content string) (FrontMatter, error) {
	attributes := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(content), &attributes)
	if err != nil {
		return FrontMatter{}, errors.Wrapf(err, "unable to parse yaml content in front-matter of document")
	}
	log.Debugf("Initialized a new FrontMatter with attributes: %+v", attributes)
	return FrontMatter{Content: attributes}, nil
}

// ------------------------------------------
// Sections
// ------------------------------------------

// Section the structure for a section
type Section struct {
	Level    int
	Title    SectionTitle
	Elements []interface{}
}

// NewSection initializes a new `Section` from the given section title and elements
func NewSection(level int, sectionTitle SectionTitle, blocks []interface{}) (Section, error) {
	log.Debugf("initializing a new Section with %d block(s)", len(blocks))
	// elements := filterEmptyElements(blocks, filterBlankLine())
	log.Debugf("Initialized a new Section of level %d with %d block(s)", level, len(blocks))
	return Section{
		Level:    level,
		Title:    sectionTitle,
		Elements: nilSafe(blocks),
	}, nil
}

// Accept implements Visitable#Accept(Visitor)
func (s Section) Accept(v Visitor) error {
	err := v.BeforeVisit(s)
	if err != nil {
		return errors.Wrapf(err, "error while pre-visiting section")
	}
	err = v.Visit(s)
	if err != nil {
		return errors.Wrapf(err, "error while visiting section")
	}
	for _, element := range s.Elements {
		if visitable, ok := element.(Visitable); ok {
			err = visitable.Accept(v)
			if err != nil {
				return errors.Wrapf(err, "error while visiting section element")
			}
		}

	}
	err = v.AfterVisit(s)
	if err != nil {
		return errors.Wrapf(err, "error while post-visiting section")
	}
	return nil
}

// ------------------------------------------
// SectionTitle
// ------------------------------------------

// SectionTitle the structure for the section titles
type SectionTitle struct {
	Attributes ElementAttributes
	Content    InlineElements
}

// NewSectionTitle initializes a new `SectionTitle`` from the given level and content, with the optional attributes.
// In the attributes, only the ElementID is retained
func NewSectionTitle(inlineContent InlineElements, attributes []interface{}) (SectionTitle, error) {
	// counting the lenght of the 'level' value (ie, the number of `=` chars)
	attrbs := NewElementAttributes(attributes)
	// make a default id from the sectionTitle's inline content
	if _, found := attrbs[AttrID]; !found {
		replacement, err := replaceNonAlphanumerics(inlineContent, "_")
		if err != nil {
			return SectionTitle{}, errors.Wrapf(err, "unable to generate default ID while instanciating a new SectionTitle element")
		}
		attrbs[AttrID] = replacement
	}
	sectionTitle := SectionTitle{
		Attributes: attrbs,
		Content:    inlineContent,
	}
	if log.GetLevel() == log.DebugLevel {
		log.Debugf("Initialized a new SectionTitle with content %v", inlineContent)
		spew.Dump(sectionTitle)
	}
	return sectionTitle, nil
}

// ------------------------------------------
// Lists
// ------------------------------------------

// List a List
type List interface {
	// AddItems() []interface{}
}

// ListItem a list item
type ListItem interface {
	AddChild(interface{})
}

// NewList initializes a new `List` from the given content
func NewList(elements []interface{}, attributes []interface{}) (List, error) {
	log.Debugf("initializing a new List with %d element(s)", len(elements))
	buffer := make(map[reflect.Type][]ListItem)
	rootType := reflect.TypeOf(toPtr(elements[0])) // elements types will be pointers
	previousType := rootType
	stack := make([]reflect.Type, 0)
	stack = append(stack, rootType)
	for _, element := range elements {
		log.Debugf("processing list item of type %T", element)
		// val := reflect.ValueOf(element).Elem().Addr().Interface()
		item, ok := toPtr(element).(ListItem)
		if !ok {
			return nil, errors.Errorf("element of type '%T' is not a valid list item", element)
		}
		// collect all elements of the same kind and make a sub list from them
		// each time a change of type is detected, except for the root type
		currentType := reflect.TypeOf(item)
		if currentType != previousType && previousType != rootType {
			log.Debugf(" detected a switch of type when processing item of type %T: currentType=%v != previousType=%v", item, currentType, previousType)
			// change of type: make a list from the buffer[t], reset and keep iterating
			sublist, err := newList(buffer[previousType], nil)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to initialize a new sublist")
			}
			// look-up the previous item of the same type as the current type
			parentItems := buffer[currentType]
			parentItem := parentItems[len(parentItems)-1]
			parentItem.AddChild(sublist)
			buffer[previousType] = make([]ListItem, 0)
			// add element type to stack if not already found
			found := false
			for _, t := range stack {
				log.Debugf("comparing stack type %v to %v: %t", t, previousType, (t == previousType))
				if t == previousType {
					found = true
					break
				}
			}
			if !found {
				log.Debugf("adding element of type %v to stack", previousType)
				stack = append(stack, previousType)
			}
		}
		previousType = currentType
		// add item to buffer
		buffer[currentType] = append(buffer[currentType], item)
	}
	// end of processing: take into account the remainings in the buffer, by stack
	log.Debugf("end of list init: stack=%v, buffer= %v", stack, buffer)
	// process all sub lists
	for i := len(stack) - 1; i > 0; i-- {
		// skip if no item at this layer/level
		if len(buffer[stack[i]]) == 0 {
			continue
		}
		// look-up parent layer at the previous (ie, upper) level in the stack
		parentItems := buffer[stack[i-1]]
		// look-up parent in the layer
		parentItem := parentItems[len(parentItems)-1]
		// build a new list from the remaining items at the current level of the stack
		sublist, err := newList(buffer[stack[i]], nil)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to initialize a new sublist")
		}
		// add this list to the parent
		parentItem.AddChild(sublist)
	}

	// log.Debugf("end of list init: current type=%v / previous type=%v / buffer= %v", currentType, previousType, buffer)
	// finally, the top-level list
	return newList(buffer[rootType], attributes)
}

func newList(items []ListItem, attributes []interface{}) (List, error) {
	// log.Debugf("initializing a new list with %d items", len(items))
	if len(items) == 0 {
		return nil, errors.Errorf("cannot build a list from an empty slice")
	}
	switch items[0].(type) {
	case *OrderedListItem:
		return NewOrderedList(items, attributes)
	case *UnorderedListItem:
		return NewUnorderedList(items, attributes)
	case *LabeledListItem:
		return NewLabeledList(items, attributes)
	default:
		return nil, errors.Errorf("unsupported type of element as the root list: '%T'", items[0])
	}
}

// ------------------------------------------
// Ordered Lists
// ------------------------------------------

// OrderedList the structure for the Ordered Lists
type OrderedList struct {
	Attributes ElementAttributes
	Items      []OrderedListItem
}

// NumberingStyle the type of numbering for items in an ordered list
type NumberingStyle string

const (
	// UnknownNumberingStyle the default, unknown type
	UnknownNumberingStyle NumberingStyle = "unknown"
	// Arabic the arabic numbering (1, 2, 3, etc.)
	Arabic NumberingStyle = "arabic"
	// Decimal the decimal numbering (01, 02, 03, etc.)
	Decimal NumberingStyle = "decimal"
	// LowerAlpha the lower-alpha numbering (a, b, c, etc.)
	LowerAlpha NumberingStyle = "loweralpha"
	// UpperAlpha the upper-alpha numbering (A, B, C, etc.)
	UpperAlpha NumberingStyle = "upperalpha"
	// LowerRoman the lower-roman numbering (i, ii, iii, etc.)
	LowerRoman NumberingStyle = "lowerroman"
	// UpperRoman the upper-roman numbering (I, II, III, etc.)
	UpperRoman NumberingStyle = "upperroman"
	// LowerGreek the lower-greek numbering (alpha, beta, etc.)
	LowerGreek NumberingStyle = "lowergreek"
	// UpperGreek the upper-roman numbering (Alpha, Beta, etc.)
	UpperGreek NumberingStyle = "uppergreek"
)

var numberingStyles []NumberingStyle

func init() {
	numberingStyles = []NumberingStyle{Arabic, Decimal, LowerAlpha, UpperAlpha, LowerRoman, UpperRoman, LowerGreek, UpperGreek}
}

// NewOrderedList initializes a new `OrderedList` from the given content
func NewOrderedList(elements []ListItem, attributes []interface{}) (OrderedList, error) {
	log.Debugf("initializing a new OrderedList from %d element(s)...", len(elements))
	result := make([]OrderedListItem, 0)
	bufferedItemsPerLevel := make(map[int][]*OrderedListItem, 0) // buffered items for the current level
	levelPerStyle := make(map[NumberingStyle]int, 0)
	previousLevel := 0
	previousNumberingStyle := UnknownNumberingStyle
	for _, element := range elements {
		item, ok := element.(*OrderedListItem)
		if !ok {
			return OrderedList{}, errors.Errorf("element of type '%T' is not a valid unorderedlist item", element)
		}
		log.Debugf("processing list item: %v", item.Elements[0])
		if item.Level > previousLevel {
			// force the current item level to (last seen level + 1)
			item.Level = previousLevel + 1
			// log.Debugf("setting item level to %d (#1 - new level)", item.Level)
			levelPerStyle[item.NumberingStyle] = item.Level
		} else if item.NumberingStyle != previousNumberingStyle {
			// check if this numbering type was already found previously
			if level, found := levelPerStyle[item.NumberingStyle]; found {
				item.Level = level // 0-based offset in the bufferedItemsPerLevel
				// log.Debugf("setting item level to %d / %v (#2 - existing style)", item.Level, item.NumberingStyle)
			} else {
				item.Level = previousLevel + 1
				// log.Debugf("setting item level to %d (#3 - new level for numbering style %v)", item.Level, item.NumberingStyle)
				levelPerStyle[item.NumberingStyle] = item.Level
			}
		} else if item.NumberingStyle == previousNumberingStyle {
			item.Level = previousLevel
			// log.Debugf("setting item level to %d (#4 - same as previous item)", item.Level)
		}
		// log.Debugf("list item %v -> level= %d", item.Elements[0], item.Level)
		// join item *values* in the parent item when the level decreased
		if item.Level < previousLevel {
			parentLayer := bufferedItemsPerLevel[previousLevel-2]
			parentItem := parentLayer[len(parentLayer)-1]
			log.Debugf("moving buffered items at level %d (%v) in parent (%v) ", previousLevel, bufferedItemsPerLevel[previousLevel-1][0].NumberingStyle, parentItem.NumberingStyle)
			childList := toOrderedList(bufferedItemsPerLevel[previousLevel-1])
			parentItem.Elements = append(parentItem.Elements, childList)
			// clear the previously buffered items at level 'previousLevel'
			delete(bufferedItemsPerLevel, previousLevel-1)
		}
		// new level of element: put it in the buffer
		if item.Level > len(bufferedItemsPerLevel) {
			// log.Debugf("initializing a new level of list items: %d", item.Level)
			bufferedItemsPerLevel[item.Level-1] = make([]*OrderedListItem, 0)
		}
		// append item to buffer of its level
		log.Debugf("adding list item %v in the current buffer at level %d", item.Elements[0], item.Level)
		bufferedItemsPerLevel[item.Level-1] = append(bufferedItemsPerLevel[item.Level-1], item)
		previousLevel = item.Level
		previousNumberingStyle = item.NumberingStyle
	}
	log.Debugf("processing the rest of the buffer...")
	// clear the remaining buffer and get the result in the reverse order of levels
	for level := len(bufferedItemsPerLevel) - 1; level >= 0; level-- {
		items := bufferedItemsPerLevel[level]
		// top-level items
		if level == 0 {
			for idx, item := range items {
				// set the position
				// log.Debugf("setting item #%d position to %d+%d", (idx + 1), items[0].Position, idx)
				item.Position = items[0].Position + idx
				result = append(result, *item)
			}
		} else {
			childList := toOrderedList(items)
			parentLayer := bufferedItemsPerLevel[level-1]
			parentItem := parentLayer[len(parentLayer)-1]
			parentItem.Elements = append(parentItem.Elements, childList)
		}
	}

	return OrderedList{
		Attributes: mergeAttributes(attributes),
		Items:      result,
	}, nil
}

func toOrderedList(items []*OrderedListItem) OrderedList {
	result := OrderedList{
		Attributes: ElementAttributes{}, // avoid nil `attributes`
	}
	// set the position and numbering style based on the optional attributes of the first item
	if len(items) == 0 {
		return result
	}
	items[0].applyAttributes()
	for idx, item := range items {
		// log.Debugf("setting item #%d position to %d+%d", (idx + 1), bufferedItemsPerLevel[previousLevel-1][0].Position, idx)
		item.Position = items[0].Position + idx
		item.NumberingStyle = items[0].NumberingStyle
		result.Items = append(result.Items, *item)
	}
	return result
}

// OrderedListItem the structure for the ordered list items
type OrderedListItem struct {
	Level          int
	Position       int
	NumberingStyle NumberingStyle
	Elements       []interface{}
	Attributes     ElementAttributes
}

// making sure that the `ListItem` interface is implemented by `OrderedListItem`
var _ ListItem = &OrderedListItem{}

// NewOrderedListItem initializes a new `orderedListItem` from the given content
func NewOrderedListItem(prefix OrderedListItemPrefix, elements []interface{}, attributes []interface{}) (OrderedListItem, error) {
	log.Debugf("initializing a new OrderedListItem with attributes %v", attributes)
	p := 1 // default position
	return OrderedListItem{
		NumberingStyle: prefix.NumberingStyle,
		Level:          prefix.Level,
		Position:       p,
		Elements:       elements,
		Attributes:     mergeAttributes(attributes),
	}, nil
}

// AddChild appends the given item to the content of this OrderedListItem
func (i *OrderedListItem) AddChild(item interface{}) {
	log.Debugf("Adding item %v to %v", item, i.Elements)
	i.Elements = append(i.Elements, item)
}

func (i *OrderedListItem) applyAttributes() error {
	log.Debugf("applying attributes on %[1]v: %[2]v (%[2]T)", i.Elements[0], i.Attributes)
	// numbering type override
	for _, style := range numberingStyles {
		if _, ok := i.Attributes[string(style)]; ok {
			i.NumberingStyle = style
			break
		}
	}
	// numbering offset
	if start, ok := i.Attributes["start"]; ok {
		if start, ok := start.(string); ok {
			s, err := strconv.ParseInt(start, 10, 64)
			if err != nil {
				return errors.Wrapf(err, "unable to parse 'start' value %v", start)
			}
			i.Position = int(s)
		}
	}
	log.Debugf("applied attributes on %v: position=%d, numbering=%v", i.Elements[0], i.Position, i.NumberingStyle)
	return nil
}

// OrderedListItemPrefix the prefix used to construct an OrderedListItem
type OrderedListItemPrefix struct {
	NumberingStyle NumberingStyle
	Level          int
}

// NewOrderedListItemPrefix initializes a new OrderedListItemPrefix
func NewOrderedListItemPrefix(s NumberingStyle, l int) (OrderedListItemPrefix, error) {
	return OrderedListItemPrefix{
		NumberingStyle: s,
		Level:          l,
	}, nil
}

// ------------------------------------------
// Unordered Lists
// ------------------------------------------

// UnorderedList the structure for the Unordered Lists
type UnorderedList struct {
	Attributes ElementAttributes
	Items      []UnorderedListItem
}

// NewUnorderedList initializes a new `UnorderedList` from the given content
func NewUnorderedList(elements []ListItem, attributes []interface{}) (UnorderedList, error) {
	log.Debugf("initializing a new UnorderedList from %d element(s)...", len(elements))
	result := make([]UnorderedListItem, 0)
	bufferedItemsPerLevel := make(map[int][]*UnorderedListItem, 0) // buffered items for the current level
	levelPerStyle := make(map[BulletStyle]int, 0)
	previousLevel := 0
	previousBulletStyle := UnknownBulletStyle
	for _, element := range elements {
		item, ok := element.(*UnorderedListItem)
		if !ok {
			return UnorderedList{}, errors.Errorf("element of type '%T' is not a valid unorderedlist item", element)
		}
		if item.Level > previousLevel {
			// force the current item level to (last seen level + 1)
			item.adjustBulletStyle(previousBulletStyle)
			item.Level = previousLevel + 1
			levelPerStyle[item.BulletStyle] = item.Level
		} else if item.BulletStyle != previousBulletStyle {
			if level, found := levelPerStyle[item.BulletStyle]; found {
				item.Level = level
			} else {
				item.Level = previousLevel + 1
				levelPerStyle[item.BulletStyle] = item.Level
			}
		} else if item.BulletStyle == previousBulletStyle {
			// adjust level on previous item of same style (in case the level
			// of the latter has been adjusted before)
			item.Level = previousLevel
		}
		log.Debugf("Processing list item of level %d: %v", item.Level, item.Elements[0])
		// join item *values* in the parent item when the level decreased
		if item.Level < previousLevel {
			// merge previous levels in parents.
			// eg: when reaching `list item 2`, the level 3 items must be merged into the level 2 item, which must
			// be itself merged in the level 1 item:
			// * list item 1
			// ** nested list item
			// *** nested nested list item 1
			// *** nested nested list item 2
			// * list item 2
			for l := previousLevel; l > item.Level; l-- {
				log.Debugf("merging previously buffered items at level '%d' in parent", l)
				parentLayer := bufferedItemsPerLevel[l-2]
				parentItem := parentLayer[len(parentLayer)-1]
				childList := UnorderedList{
					Attributes: ElementAttributes{}, // avoid nil `attributes`
				}
				for _, i := range bufferedItemsPerLevel[l-1] {
					childList.Items = append(childList.Items, *i)
				}
				parentItem.Elements = append(parentItem.Elements, childList)
				// clear the previously buffered items at level 'previousLevel'
				delete(bufferedItemsPerLevel, l-1)
			}
		}
		// new level of element: put it in the buffer
		if item.Level > len(bufferedItemsPerLevel) {
			log.Debugf("initializing a new level of list items: %d", item.Level)
			bufferedItemsPerLevel[item.Level-1] = make([]*UnorderedListItem, 0)
		}
		// append item to buffer of its level
		log.Debugf("adding list item %v in the current buffer", item.Elements[0])
		bufferedItemsPerLevel[item.Level-1] = append(bufferedItemsPerLevel[item.Level-1], item)
		previousLevel = item.Level
		previousBulletStyle = item.BulletStyle
	}
	log.Debugf("processing the rest of the buffer: %v", bufferedItemsPerLevel)
	// clear the remaining buffer and get the result in the reverse order of levels
	for level := len(bufferedItemsPerLevel) - 1; level >= 0; level-- {
		items := bufferedItemsPerLevel[level]
		// top-level items
		if level == 0 {
			for _, item := range items {
				result = append(result, *item)
			}
		} else {
			childList := UnorderedList{
				Attributes: ElementAttributes{}, // avoid nil `attributes`
			}
			for _, item := range items {
				childList.Items = append(childList.Items, *item)
			}
			parentLayer := bufferedItemsPerLevel[level-1]
			parentItem := parentLayer[len(parentLayer)-1]
			parentItem.Elements = append(parentItem.Elements, childList)
		}
	}

	return UnorderedList{
		Attributes: mergeAttributes(attributes),
		Items:      result,
	}, nil
}

// UnorderedListItem the structure for the unordered list items
type UnorderedListItem struct {
	Level       int
	BulletStyle BulletStyle
	Elements    []interface{}
}

// NewUnorderedListItem initializes a new `UnorderedListItem` from the given content
func NewUnorderedListItem(prefix UnorderedListItemPrefix, elements []interface{}) (UnorderedListItem, error) {
	log.Debugf("initializing a new UnorderedListItem...")
	// log.Debugf("initializing a new UnorderedListItem with '%d' lines (%T) and input level '%d'", len(elements), elements, lvl.Len())
	return UnorderedListItem{
		Level:       prefix.Level,
		BulletStyle: prefix.BulletStyle,
		Elements:    elements,
	}, nil
}

// AddChild appends the given item to the content of this UnorderedListItem
func (i *UnorderedListItem) AddChild(item interface{}) {
	i.Elements = append(i.Elements, item)
}

// adjustBulletStyle
func (i *UnorderedListItem) adjustBulletStyle(p BulletStyle) {
	n := i.BulletStyle.nextLevelStyle(p)
	log.Debugf("adjusting bullet style for item with level '%v' to '%v' (previously processed/parent level: '%v')", i.BulletStyle, p, n)
	i.BulletStyle = n
}

// BulletStyle the type of bullet for items in an unordered list
type BulletStyle string

const (
	// UnknownBulletStyle the default, unknown type
	UnknownBulletStyle BulletStyle = "unkwown"
	// Dash an unordered item can begin with a single dash
	Dash BulletStyle = "dash"
	// OneAsterisk an unordered item marked with a single asterisk
	OneAsterisk BulletStyle = "1asterisk"
	// TwoAsterisks an unordered item marked with two asterisks
	TwoAsterisks BulletStyle = "2asterisks"
	// ThreeAsterisks an unordered item marked with three asterisks
	ThreeAsterisks BulletStyle = "3asterisks"
	// FourAsterisks an unordered item marked with four asterisks
	FourAsterisks BulletStyle = "4asterisks"
	// FiveAsterisks an unordered item marked with five asterisks
	FiveAsterisks BulletStyle = "5asterisks"
)

// nextLevelStyle returns the BulletStyle for the next level:
// `-` -> `*`
// `*` -> `**`
// `**` -> `***`
// `***` -> `****`
// `****` -> `*****`
// `*****` -> `-`

func (b BulletStyle) nextLevelStyle(p BulletStyle) BulletStyle {
	switch p {
	case Dash:
		return OneAsterisk
	case OneAsterisk:
		return TwoAsterisks
	case TwoAsterisks:
		return ThreeAsterisks
	case ThreeAsterisks:
		return FourAsterisks
	case FourAsterisks:
		return FiveAsterisks
	case FiveAsterisks:
		return Dash
	}
	// default, return the level itself
	return b
}

// UnorderedListItemPrefix the prefix used to construct an UnorderedListItem
type UnorderedListItemPrefix struct {
	BulletStyle BulletStyle
	Level       int
}

// NewUnorderedListItemPrefix initializes a new UnorderedListItemPrefix
func NewUnorderedListItemPrefix(s BulletStyle, l int) (UnorderedListItemPrefix, error) {
	return UnorderedListItemPrefix{
		BulletStyle: s,
		Level:       l,
	}, nil
}

// NewListItemContent initializes a new `UnorderedListItemContent`
func NewListItemContent(content []interface{}) ([]interface{}, error) {
	// log.Debugf("initializing a new ListItemContent with %d line(s)", len(content))
	elements := make([]interface{}, 0)
	for _, element := range content {
		// log.Debugf("Processing line element of type %T", element)
		switch element := element.(type) {
		case []interface{}:
			for _, e := range element {
				// if e, ok := e.(interface{}); ok {
				elements = append(elements, e)
				// }
			}
		case interface{}:
			elements = append(elements, element)
		}
	}
	// log.Debugf("Initialized a new ListItemContent with %d elements(s)", len(elements))
	// no need to return an empty ListItemContent
	if len(elements) == 0 {
		return nil, nil
	}
	return elements, nil
}

// ListItemContinuation a list item continuation
type ListItemContinuation struct {
}

// NewListItemContinuation returns a new ListItemContinuation
func NewListItemContinuation() (ListItemContinuation, error) {
	return ListItemContinuation{}, nil
}

// ------------------------------------------
// Labeled List
// ------------------------------------------

// LabeledList the structure for the Labeled Lists
type LabeledList struct {
	Attributes ElementAttributes
	Items      []LabeledListItem
}

// NewLabeledList initializes a new `LabeledList` from the given content
func NewLabeledList(elements []ListItem, attributes []interface{}) (LabeledList, error) {
	log.Debugf("initializing a new LabeledList from %d elements", len(elements))
	items := make([]LabeledListItem, 0)
	for _, element := range elements {
		if item, ok := element.(*LabeledListItem); ok {
			items = append(items, *item)
		}
	}
	log.Debugf("Initialized a new LabeledList with %d root item(s)", len(items))
	return LabeledList{
		Attributes: mergeAttributes(attributes),
		Items:      items,
	}, nil
}

// LabeledListItem an item in a labeled
type LabeledListItem struct {
	Term     string
	Elements []interface{}
}

// NewLabeledListItem initializes a new LabeledListItem
func NewLabeledListItem(term string, elements []interface{}) (LabeledListItem, error) {
	log.Debugf("initializing a new LabeledListItem with %d elements (%T)", len(elements), elements)
	return LabeledListItem{
		Term:     term,
		Elements: elements,
	}, nil
}

// AddChild appends the given item to the content of this LabeledListItem
func (i *LabeledListItem) AddChild(item interface{}) {
	log.Debugf("Adding item %v to %v", item, i.Elements)
	i.Elements = append(i.Elements, item)
}

// making sure that the `ListItem` interface is implemented by `LabeledListItem`
var _ ListItem = &LabeledListItem{}

// ------------------------------------------
// Paragraph
// ------------------------------------------

// Paragraph the structure for the paragraphs
type Paragraph struct {
	Attributes ElementAttributes
	Lines      []InlineElements
}

// NewParagraph initializes a new `Paragraph`
func NewParagraph(lines []interface{}, attributes []interface{}) (Paragraph, error) {
	log.Debugf("initializing a new Paragraph with %d line(s)", len(lines))
	attrbs := NewElementAttributes(attributes)
	elements := make([]InlineElements, 0)
	for _, line := range lines {
		if l, ok := line.(InlineElements); ok {
			log.Debugf(" processing paragraph line of type %T", line)
			// if len(l) > 0 {
			elements = append(elements, l)
			// }
		} else {
			return Paragraph{}, errors.Errorf("unsupported paragraph line of type %[1]T: %[1]v", line)
		}

	}
	log.Debugf("generated a paragraph with %d line(s): %v", len(elements), elements)
	return Paragraph{
		Attributes: attrbs,
		Lines:      elements,
	}, nil
}

// NewAdmonitionParagraph returns a new Paragraph with an extra admonition attribute
func NewAdmonitionParagraph(lines []interface{}, admonitionKind AdmonitionKind, attributes []interface{}) (Paragraph, error) {
	p, err := NewParagraph(lines, attributes)
	if err != nil {
		return p, err
	}
	p.Attributes[AttrAdmonitionKind] = admonitionKind
	return p, nil
}

// ------------------------------------------
// Admonitions
// ------------------------------------------

// AdmonitionKind the type of admonition
type AdmonitionKind string

const (
	// Tip the 'TIP' type of admonition
	Tip AdmonitionKind = "tip"
	// Note the 'NOTE' type of admonition
	Note AdmonitionKind = "note"
	// Important the 'IMPORTANT' type of admonition
	Important AdmonitionKind = "important"
	// Warning the 'WARNING' type of admonition
	Warning AdmonitionKind = "warning"
	// Caution the 'CAUTION' type of admonition
	Caution AdmonitionKind = "caution"
	// Unknown is the zero value for admonition kind
	Unknown AdmonitionKind = ""
)

// ------------------------------------------
// InlineElements
// ------------------------------------------

// InlineElements the structure for the lines in paragraphs
type InlineElements []interface{}

// NewInlineElements initializes a new `InlineElements` from the given values
func NewInlineElements(elements ...interface{}) (InlineElements, error) {
	result := mergeElements(elements...)
	return result, nil
}

// Accept implements Visitable#Accept(Visitor)
func (e InlineElements) Accept(v Visitor) error {
	err := v.BeforeVisit(e)
	if err != nil {
		return errors.Wrapf(err, "error while pre-visiting inline content")
	}
	err = v.Visit(e)
	if err != nil {
		return errors.Wrapf(err, "error while visiting inline content")
	}
	for _, element := range e {
		if visitable, ok := element.(Visitable); ok {
			err = visitable.Accept(v)
			if err != nil {
				return errors.Wrapf(err, "error while visiting inline content element")
			}
		}
	}
	err = v.AfterVisit(e)
	if err != nil {
		return errors.Wrapf(err, "error while post-visiting sectionTitle")
	}
	return nil
}

// ------------------------------------------
// Cross References
// ------------------------------------------

// CrossReference the struct for Cross References
type CrossReference struct {
	ID    string
	Label string
}

// NewCrossReference initializes a new `CrossReference` from the given ID
func NewCrossReference(id string, label interface{}) (CrossReference, error) {
	log.Debugf("initializing a new CrossReference with ID=%s", id)
	var l string
	if label, ok := label.(string); ok {
		l = apply(label, strings.TrimSpace)
	}
	return CrossReference{
		ID:    id,
		Label: l,
	}, nil
}

// ------------------------------------------
// Images
// ------------------------------------------

const (
	// AttrImageAlt the image `alt` attribute
	AttrImageAlt string = "alt"
	// AttrImageWidth the image `width` attribute
	AttrImageWidth string = "width"
	// AttrImageHeight the image `height` attribute
	AttrImageHeight string = "height"
	// AttrImageTitle the image `title` attribute
	AttrImageTitle string = "title"
)

// BlockImage the structure for the block images
type BlockImage struct {
	Path       string
	Attributes ElementAttributes
}

// NewBlockImage initializes a new `BlockImage`
func NewBlockImage(path string, attributes []interface{}, inlineAttributes ElementAttributes) (BlockImage, error) {
	allAttributes := mergeAttributes(attributes)
	for k, v := range inlineAttributes {
		allAttributes[k] = v
	}
	if alt, found := allAttributes[AttrImageAlt]; !found || alt == "" {
		_, filename := filepath.Split(path)
		log.Debugf("adding alt based on filename '%s'", filename)
		ext := filepath.Ext(filename)
		if ext != "" {
			allAttributes[AttrImageAlt] = strings.TrimRight(filename, fmt.Sprintf(".%s", ext))
		} else {
			allAttributes[AttrImageAlt] = filename
		}
	}
	return BlockImage{
		Path:       path,
		Attributes: allAttributes,
	}, nil
}

// InlineImage the structure for the inline image macros
type InlineImage struct {
	Path       string
	Attributes ElementAttributes
}

// NewInlineImage initializes a new `InlineImage` (similar to BlockImage, but without attributes)
func NewInlineImage(path string, attributes ElementAttributes) (InlineImage, error) {
	if alt, found := attributes[AttrImageAlt]; !found || alt == "" {
		_, filename := filepath.Split(path)
		log.Debugf("adding alt based on filename '%s'", filename)
		ext := filepath.Ext(filename)
		if ext != "" {
			attributes[AttrImageAlt] = strings.TrimRight(filename, fmt.Sprintf(".%s", ext))
		} else {
			attributes[AttrImageAlt] = filename
		}
	}
	return InlineImage{
		Path:       path,
		Attributes: attributes,
	}, nil
}

// ImageMacro the structure for the block image macros
type ImageMacro struct {
	Path       string
	Attributes ElementAttributes
}

// NewImageMacro initializes a new `ImageMacro`
func NewImageMacro(path string, attributes ElementAttributes) (ImageMacro, error) {
	// use the image filename without the extension as the default `alt` attribute
	log.Debugf("processing alt: '%s'", attributes[AttrImageAlt])
	if attributes[AttrImageAlt] == "" {
		_, filename := filepath.Split(path)
		log.Debugf("adding alt based on filename '%s'", filename)
		ext := filepath.Ext(filename)
		if ext != "" {
			attributes[AttrImageAlt] = strings.TrimRight(filename, fmt.Sprintf(".%s", ext))
		} else {
			attributes[AttrImageAlt] = filename
		}
	}
	return ImageMacro{
		Path:       path,
		Attributes: attributes,
	}, nil
}

// NewImageAttributes returns a map of image attributes, some of which have implict keys (`alt`, `width` and `height`)
func NewImageAttributes(alt, width, height interface{}, otherAttrs []interface{}) (ElementAttributes, error) {
	result := ElementAttributes{}
	var altStr, widthStr, heightStr string
	if alt, ok := alt.(string); ok {
		altStr = apply(alt, strings.TrimSpace)
	}
	if width, ok := width.(string); ok {
		widthStr = apply(width, strings.TrimSpace)
	}
	if height, ok := height.(string); ok {
		heightStr = apply(height, strings.TrimSpace)
	}
	result[AttrImageAlt] = altStr
	result[AttrImageWidth] = widthStr
	result[AttrImageHeight] = heightStr
	for _, otherAttr := range otherAttrs {
		if otherAttr, ok := otherAttr.(ElementAttributes); ok {
			for k, v := range otherAttr {
				result[k] = v
			}
		}
	}
	return result, nil
}

// ------------------------------------------
// Delimited blocks
// ------------------------------------------

// DelimitedBlock the structure for the delimited blocks
type DelimitedBlock struct {
	Attributes ElementAttributes
	Elements   []interface{}
}

// Substitution the substituion group to apply when initializing a delimited block
type Substitution func([]interface{}) ([]interface{}, error)

// None returns the content as-is, but nil-safe
func None(content []interface{}) ([]interface{}, error) {
	return nilSafe(content), nil
}

// Verbatim the verbatim substitution: the given content is converted into an array of strings.
func Verbatim(content []interface{}) ([]interface{}, error) {
	result := make([]interface{}, len(content))
	for i, c := range content {
		if c, ok := c.(string); ok {
			c = apply(c, func(s string) string {
				return strings.TrimRight(c, "\n\r")
			})
			result[i] = NewStringElement(c)
		}
	}
	return result, nil
}

// NewDelimitedBlock initializes a new `DelimitedBlock` of the given kind with the given content
func NewDelimitedBlock(kind BlockKind, content []interface{}, attributes []interface{}, substitution Substitution) (DelimitedBlock, error) {
	log.Debugf("Initializing a new DelimitedBlock of kind '%v'", kind)
	attrbs := NewElementAttributes(attributes)
	if _, found := attrbs[AttrKind]; !found {
		attrbs[AttrKind] = kind
	}
	elements, err := substitution(content)
	if err != nil {
		return DelimitedBlock{}, errors.Wrapf(err, "failed to initialize a new delimited block")
	}
	return DelimitedBlock{
		Attributes: attrbs,
		Elements:   elements,
	}, nil
}

// ------------------------------------------
// Tables
// ------------------------------------------

// Table the structure for the tables
type Table struct {
	Attributes ElementAttributes
	Header     TableLine
	Lines      []TableLine
}

// NewTable initializes a new table with the given lines and attributes
func NewTable(header interface{}, lines []interface{}, attributes []interface{}) (Table, error) {
	t := Table{
		Attributes: NewElementAttributes(attributes),
	}
	columnsPerLine := -1 // unknown until first "line" is processed
	if header, ok := header.(TableLine); ok {
		t.Header = header
		columnsPerLine = len(header.Cells)
	}
	// need to regroup columns of all lines, they dispatch on lines
	cells := make([]InlineElements, 0)
	for _, l := range lines {
		if l, ok := l.(TableLine); ok {
			// if no header line was set, inspect the first line to determine the number of columns per line
			if columnsPerLine == -1 {
				columnsPerLine = len(l.Cells)
			}
			cells = append(cells, l.Cells...)
		}
	}
	t.Lines = make([]TableLine, 0)
	if len(lines) > 0 {
		log.Debugf("buffered %d columns for the table", len(cells))
		l := TableLine{
			Cells: make([]InlineElements, columnsPerLine),
		}
		for i, c := range cells {
			log.Debugf("adding cell with content '%v' in table line at offset %d", c, (i % columnsPerLine))
			l.Cells[i%columnsPerLine] = c
			if (i+1)%columnsPerLine == 0 { // switch to next line
				log.Debugf("adding line with content '%v' in table", l)
				t.Lines = append(t.Lines, l)
				l = TableLine{
					Cells: make([]InlineElements, columnsPerLine),
				}
			}
		}
	}
	log.Debugf("initialized a new table with %d line(s)", len(lines))
	return t, nil
}

// TableLine a table line is made of columns, each column being a group of InlineElements (to support quoted text, etc.)
type TableLine struct {
	Cells []InlineElements
}

// NewTableLine initializes a new TableLine with the given columns
func NewTableLine(columns []interface{}) (TableLine, error) {
	c := make([]InlineElements, 0)
	for _, column := range columns {
		if e, ok := column.(InlineElements); ok {
			c = append(c, e)
		} else {
			log.Debugf("unsupported element of type %T", column)
		}
	}
	log.Debugf("initialized a new table line with %d columns", len(c))
	return TableLine{
		Cells: c,
	}, nil
}

// ------------------------------------------
// Literal blocks
// ------------------------------------------

// LiteralBlock the structure for the literal blocks
type LiteralBlock struct {
	Content string
}

// NewLiteralBlock initializes a new `DelimitedBlock` of the given kind with the given content,
// along with the given sectionTitle spaces
func NewLiteralBlock(content string) (LiteralBlock, error) {
	// concatenates the spaces with the actual content in a single 'stringified' value
	// log.Debugf("initializing a new LiteralBlock with spaces='%v' and content=`%v`", spaces, content)
	// remove "\n" or "\r\n", depending on the OS.
	blockContent := apply(content,
		func(s string) string {
			return strings.TrimRight(s, "\n\r")
		},
	)
	log.Debugf("Initialized a new LiteralBlock with content=`%s`", blockContent)
	return LiteralBlock{
		Content: blockContent,
	}, nil
}

// ------------------------------------------
// Comments
// ------------------------------------------

// SingleLineComment a single line comment
type SingleLineComment struct {
	Content string
}

// NewSingleLineComment initializes a new single line content
func NewSingleLineComment(content string) (SingleLineComment, error) {
	log.Debugf("initializing a single line comment with content: '%s'", content)
	return SingleLineComment{
		Content: content,
	}, nil
}

// ------------------------------------------
// Elements attributes
// ------------------------------------------

const (
	// AttrID the key to retrieve the ID in the element attributes
	AttrID string = "id"
	// AttrTitle the key to retrieve the title in the element attributes
	AttrTitle string = "title"
	// AttrRole the key to retrieve the role in the element attributes
	AttrRole string = "role"
	// AttrLink the key to retrieve the link in the element attributes
	AttrLink string = "link"
	// AttrAdmonitionKind the key to retrieve the kind of admonition in the element attributes, if a "masquerade" is used
	AttrAdmonitionKind string = "admonitionKind"
	// AttrQuoteAuthor attribute for the author of a verse
	AttrQuoteAuthor string = "quoteAuthor"
	// AttrQuoteTitle attribute for the title of a verse
	AttrQuoteTitle string = "quoteTitle"
)

// ElementAttributes is a map[string]interface{} with some utility methods
type ElementAttributes map[string]interface{}

// GetAsString returns the value of the key as a string, or empty string if the key did not exist
func (a ElementAttributes) GetAsString(key string) string {
	if v, ok := a[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// NewElementAttributes retrieves the ElementID, ElementTitle and ElementLink from the given slice of attributes
func NewElementAttributes(attributes []interface{}) ElementAttributes {
	attrbs := make(ElementAttributes)
	for _, attrb := range attributes {
		log.Debugf("processing attribute %[1]v (%[1]T)", attrb)
		switch attrb := attrb.(type) {
		case ElementAttributes:
			// TODO: warn if attribute already exists and is overridden
			for k, v := range attrb {
				attrbs[k] = v
			}
		case map[string]interface{}:
			// TODO: warn if attribute already exists and is overridden
			for k, v := range attrb {
				attrbs[k] = v
			}
		case nil:
			// ignore
		default:
			log.Warnf("Unexpected attributes: %T", attrb)
		}
	}
	return attrbs
}

// NewElementID initializes a new attribute map with a single entry for the ID using the given value
func NewElementID(id string) (ElementAttributes, error) {
	log.Debugf("initializing a new ElementID with ID=%s", id)
	return ElementAttributes{AttrID: id}, nil
}

// NewElementTitle initializes a new attribute map with a single entry for the title using the given value
func NewElementTitle(title string) (ElementAttributes, error) {
	log.Debugf("initializing a new ElementTitle with content=%s", title)
	return ElementAttributes{
		AttrTitle: title,
	}, nil
}

// NewElementRole initializes a new attribute map with a single entry for the title using the given value
func NewElementRole(role string) (ElementAttributes, error) {
	log.Debugf("initializing a new ElementRole with content=%s", role)
	return ElementAttributes{
		AttrRole: role,
	}, nil
}

// NewAdmonitionAttribute initializes a new attribute map with a single entry for the admonition kind using the given value
func NewAdmonitionAttribute(k AdmonitionKind) (ElementAttributes, error) {
	return ElementAttributes{AttrAdmonitionKind: k}, nil
}

// NewAttributeGroup initializes a group of attributes from the given generic attributes.
func NewAttributeGroup(attributes []interface{}) (ElementAttributes, error) {
	// log.Debugf("initializing a new AttributeGroup with %v", attributes)
	result := make(ElementAttributes)
	for _, a := range attributes {
		log.Debugf("processing attribute group element of type %T", a)
		if a, ok := a.(ElementAttributes); ok {
			for k, v := range a {
				result[k] = v
			}
		} else {
			return result, errors.Errorf("unable to process element of type '%[1]T': '%[1]s'", a)
		}
	}
	// log.Debugf("Initialized a new AttributeGroup: %v", result)
	return result, nil
}

// NewGenericAttribute initializes a new ElementAttribute from the given key and optional value
func NewGenericAttribute(key string, value interface{}) (ElementAttributes, error) {
	result := make(map[string]interface{})
	k := apply(key,
		// remove surrounding quotes
		func(s string) string {
			return strings.Trim(s, "\"")
		},
		strings.TrimSpace)
	if value, ok := value.(string); ok {
		v := apply(value,
			// remove surrounding quotes
			func(s string) string {
				return strings.Trim(s, "\"")
			},
			strings.TrimSpace)
		result[k] = v
	} else {
		result[k] = nil
	}
	// log.Debugf("Initialized a new ElementAttributes: %v", result)
	return result, nil
}

// NewQuoteAttributes initializes a new map of attributes for a verse paragraph
func NewQuoteAttributes(kind, author, title string) (map[string]interface{}, error) {
	result := make(map[string]interface{}, 3)
	switch kind {
	case "verse":
		result[AttrKind] = Verse
	default:
		result[AttrKind] = Quote
	}
	result[AttrQuoteAuthor] = strings.TrimSpace(author)
	result[AttrQuoteTitle] = strings.TrimSpace(title)
	log.Debugf("initialized new quote attributes: %v", result)
	return result, nil
}

// ------------------------------------------
// StringElement
// ------------------------------------------

// StringElement the structure for strings
type StringElement struct {
	Content string
}

// NewStringElement initializes a new `StringElement` from the given content
func NewStringElement(content string) StringElement {
	return StringElement{Content: content}
}

// Accept implements Visitable#Accept(Visitor)
func (s StringElement) Accept(v Visitor) error {
	err := v.BeforeVisit(s)
	if err != nil {
		return errors.Wrapf(err, "error while pre-visiting string element")
	}
	err = v.Visit(s)
	if err != nil {
		return errors.Wrapf(err, "error while visiting string element")
	}
	err = v.AfterVisit(s)
	if err != nil {
		return errors.Wrapf(err, "error while post-visiting string element")
	}
	return nil
}

// ------------------------------------------
// Quoted text
// ------------------------------------------

// QuotedText the structure for quoted text
type QuotedText struct {
	Attributes ElementAttributes
	Elements   InlineElements
}

// QuotedTextKind the type for
type QuotedTextKind int

const (
	// Bold bold quoted text
	Bold QuotedTextKind = iota
	// Italic italic quoted text
	Italic
	// Monospace monospace quoted text
	Monospace
)

// NewQuotedText initializes a new `QuotedText` from the given kind and content
func NewQuotedText(kind QuotedTextKind, content []interface{}) (QuotedText, error) {
	elements := mergeElements(content...)
	if log.GetLevel() == log.DebugLevel {
		log.Debugf("Initialized a new QuotedText with %d elements:", len(elements))
		spew.Dump(elements)
	}
	return QuotedText{
		Attributes: map[string]interface{}{AttrKind: kind},
		Elements:   elements,
	}, nil
}

// Accept implements Visitable#Accept(Visitor)
func (t QuotedText) Accept(v Visitor) error {
	err := v.BeforeVisit(t)
	if err != nil {
		return errors.Wrapf(err, "error while pre-visiting quoted text")
	}
	err = v.Visit(t)
	if err != nil {
		return errors.Wrapf(err, "error while visiting quoted text")
	}
	for _, element := range t.Elements {
		if visitable, ok := element.(Visitable); ok {
			err := visitable.Accept(v)
			if err != nil {
				return errors.Wrapf(err, "error while visiting quoted text element")
			}
		}
	}
	err = v.AfterVisit(t)
	if err != nil {
		return errors.Wrapf(err, "error while post-visiting quoted text")
	}
	return nil
}

// ------------------------------------------------------
// Escaped Quoted Text (i.e., with substitution prevention)
// ------------------------------------------------------

// NewEscapedQuotedText returns a new InlineElements where the nested elements are preserved (ie, substituted as expected)
func NewEscapedQuotedText(backslashes string, punctuation string, content []interface{}) ([]interface{}, error) {
	backslashesStr := apply(backslashes,
		func(s string) string {
			// remove the number of back-slashes that match the length of the punctuation. Eg: `\*` or `\\**`, but keep extra back-slashes
			if len(s) > len(punctuation) {
				return s[len(punctuation):]
			}
			return ""
		})
	return []interface{}{backslashesStr, punctuation, content, punctuation}, nil
}

// ------------------------------------------
// Passthrough
// ------------------------------------------

// Passthrough the structure for Passthroughs
type Passthrough struct {
	Kind     PassthroughKind
	Elements InlineElements
}

// PassthroughKind the kind of passthrough
type PassthroughKind int

const (
	// SinglePlusPassthrough a passthrough with a single `+` punctuation
	SinglePlusPassthrough PassthroughKind = iota
	// TriplePlusPassthrough a passthrough with a triple `+++` punctuation
	TriplePlusPassthrough
	// PassthroughMacro a passthrough with the `pass:[]` macro
	PassthroughMacro
)

// NewPassthrough returns a new passthrough
func NewPassthrough(kind PassthroughKind, elements []interface{}) (Passthrough, error) {
	return Passthrough{
		Kind:     kind,
		Elements: mergeElements(elements...),
	}, nil

}

// ------------------------------------------
// BlankLine
// ------------------------------------------

// BlankLine the structure for the empty lines, which are used to separate logical blocks
type BlankLine struct {
}

// NewBlankLine initializes a new `BlankLine`
func NewBlankLine() (BlankLine, error) {
	// log.Debug("initializing a new BlankLine")
	return BlankLine{}, nil
}

// ------------------------------------------
// Links
// ------------------------------------------

// Link the structure for the external links
type Link struct {
	URL        string
	Attributes ElementAttributes
}

// NewLink initializes a new `Link`
func NewLink(url string, attributes ElementAttributes) (Link, error) {
	// init attributes with empty 'text' attribute
	if attributes == nil {
		attributes = map[string]interface{}{
			AttrLinkText: "",
		}
	}
	return Link{
		URL:        url,
		Attributes: attributes,
	}, nil
}

// Text returns the `text` value for the Link,
func (l Link) Text() string {
	if text, ok := l.Attributes[AttrLinkText].(string); ok {
		return text
	}
	return ""
}

// AttrLinkText the link `text` attribute
const AttrLinkText string = "text"

// NewLinkAttributes returns a map of image attributes, some of which have implict keys (`text`)
func NewLinkAttributes(text interface{}, otherAttrs []interface{}) (ElementAttributes, error) {
	result := ElementAttributes{}
	var textStr string
	if text, ok := text.(string); ok {
		textStr = apply(text, strings.TrimSpace)
	}
	result[AttrLinkText] = textStr
	for _, otherAttr := range otherAttrs {
		if otherAttr, ok := otherAttr.(ElementAttributes); ok {
			for k, v := range otherAttr {
				result[k] = v
			}
		}
	}
	return result, nil
}
