package sgml

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/bytesparadise/libasciidoc/pkg/renderer"
	"github.com/bytesparadise/libasciidoc/pkg/types"
	"github.com/pkg/errors"
)

func (r *sgmlRenderer) renderDocumentDetails(ctx *renderer.Context) (*string, error) {
	if ctx.Attributes.Has(types.AttrAuthors) {
		authors, err := r.renderDocumentAuthorsDetails(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "error while rendering the document details")
		}
		documentDetailsBuff := &bytes.Buffer{}
		revLabel, _ := ctx.Attributes.GetAsString("version-label")
		revNumber, _ := ctx.Attributes.GetAsString("revnumber")
		revDate, _ := ctx.Attributes.GetAsString("revdate")
		revRemark, _ := ctx.Attributes.GetAsString("revremark")
		err = r.documentDetails.Execute(documentDetailsBuff, struct {
			Authors   string
			RevLabel  string
			RevNumber string
			RevDate   string
			RevRemark string
		}{
			Authors:   *authors,
			RevLabel:  revLabel,
			RevNumber: revNumber,
			RevDate:   revDate,
			RevRemark: revRemark,
		})
		if err != nil {
			return nil, errors.Wrap(err, "error while rendering the document details")
		}
		documentDetails := string(documentDetailsBuff.String()) //nolint: gosec
		return &documentDetails, nil
	}
	return nil, nil
}

func (r *sgmlRenderer) renderDocumentAuthorsDetails(ctx *renderer.Context) (*string, error) { // TODO: use  `types.DocumentAuthor` attribute in context
	authorsDetailsBuff := &strings.Builder{}
	i := 1
	for {
		var authorKey string
		var emailKey string
		var index string
		if i == 1 {
			authorKey = "author"
			emailKey = "email"
			index = ""
		} else {
			index = strconv.Itoa(i)
			authorKey = "author_" + index
			emailKey = "email_" + index
		}
		// having at least one author is the minimal requirement for document details
		if author, ok := ctx.Attributes.GetAsString(authorKey); ok {
			if i > 1 {
				authorsDetailsBuff.WriteString("\n")
			}
			email, _ := ctx.Attributes.GetAsString(emailKey)
			err := r.documentAuthorDetails.Execute(authorsDetailsBuff, struct {
				Index string
				Name  string
				Email string
			}{
				Index: index,
				Name:  author,
				Email: email,
			})
			if err != nil {
				return nil, errors.Wrap(err, "error while rendering the document author")
			}
			// if there were authors before, need to insert a `\n`
			i++
		} else {
			break
		}
	}
	result := string(authorsDetailsBuff.String()) //nolint: gosec
	return &result, nil
}
