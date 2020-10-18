package sgml

import (
	"strings"

	"github.com/bytesparadise/libasciidoc/pkg/renderer"
	"github.com/bytesparadise/libasciidoc/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (r *sgmlRenderer) renderAdmonitionBlock(ctx *renderer.Context, b types.ExampleBlock) (string, error) {
	kind, _ := b.Attributes[types.AttrAdmonitionKind].(types.AdmonitionKind)
	icon, err := r.renderIcon(ctx, types.Icon{Class: string(kind), Attributes: b.Attributes}, true)
	if err != nil {
		return "", err
	}
	result := &strings.Builder{}
	blocks := discardBlankLines(b.Elements)
	content, err := r.renderElements(ctx, blocks)
	if err != nil {
		return "", errors.Wrap(err, "unable to render admonition block content")
	}
	roles, err := r.renderElementRoles(ctx, b.Attributes)
	if err != nil {
		return "", errors.Wrap(err, "unable to render fenced block content")
	}
	err = r.admonitionBlock.Execute(result, struct {
		Context *renderer.Context
		ID      string
		Title   string
		Kind    types.AdmonitionKind
		Roles   string
		Icon    string
		Content string
	}{
		Context: ctx,
		ID:      r.renderElementID(b.Attributes),
		Kind:    kind,
		Roles:   roles,
		Title:   r.renderElementTitle(b.Attributes),
		Icon:    icon,
		Content: content,
	})
	return result.String(), err
}

func (r *sgmlRenderer) renderAdmonitionParagraph(ctx *renderer.Context, p types.Paragraph) (string, error) {
	log.Debug("rendering admonition paragraph...")
	result := &strings.Builder{}
	k, ok := p.Attributes[types.AttrAdmonitionKind].(types.AdmonitionKind)
	if !ok {
		return "", errors.Errorf("failed to render admonition with unknown kind: %T", p.Attributes[types.AttrAdmonitionKind])
	}
	icon, err := r.renderIcon(ctx, types.Icon{Class: string(k), Attributes: p.Attributes}, true)
	if err != nil {
		return "", err
	}
	content, err := r.renderLines(ctx, p.Lines)
	if err != nil {
		return "", err
	}
	roles, err := r.renderElementRoles(ctx, p.Attributes)
	if err != nil {
		return "", errors.Wrap(err, "unable to render fenced block content")
	}
	err = r.admonitionParagraph.Execute(result, struct {
		Context *renderer.Context
		ID      string
		Title   string
		Roles   string
		Icon    string
		Kind    string
		Content string
		Lines   [][]interface{}
	}{
		Context: ctx,
		ID:      r.renderElementID(p.Attributes),
		Title:   r.renderElementTitle(p.Attributes),
		Kind:    string(k),
		Roles:   roles,
		Icon:    icon,
		Content: string(content),
		Lines:   p.Lines,
	})

	return result.String(), err
}
