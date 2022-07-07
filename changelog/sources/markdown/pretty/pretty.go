package pretty

import (
	"fmt"
	"io"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
)

// Renderer is a markdown.Renderer wrapper that adds empty newlines after and before certain elements.
type Renderer struct {
	markdown.Renderer
	writtenNewline bool
}

// New returns a new Renderer.
// TODO: It would be great to be able to configure this to have customizable "before" and "after" spacing.
func New(inner markdown.Renderer) *Renderer {
	return &Renderer{
		Renderer:       inner,
		writtenNewline: true, // Initialized to true to prevent a leading newline to be written before the first header.
	}
}

// RenderNode intercepts the inner renderer's RenderNode, implementing the newline logic.
func (p *Renderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	switch {
	case entering && p.spaceBefore(node) && !p.writtenNewline:
		p.newline(w)
	case !entering && p.spaceAfter(node):
		p.newline(w)
	default:
		p.writtenNewline = false
	}

	if cb, isCodeBlock := node.(*ast.CodeBlock); isCodeBlock {
		// Hack: The default renderer just does not handle well code blocks.
		fmt.Fprintf(w, "```\n%s\n```\n", strings.TrimSpace(string(cb.Literal)))
		return ast.GoToNext
	}

	return p.Renderer.RenderNode(w, node, entering)
}

// spaceAfter returns true if an empty newline should be printer after a node.
func (p *Renderer) spaceAfter(node ast.Node) bool {
	switch node.(type) {
	case *ast.List:
		return true
	case *ast.Paragraph:
		_, isRootParagraph := node.GetParent().(*ast.Document)
		return isRootParagraph
	case *ast.CodeBlock:
		return true
	}

	return false
}

// spaceBefore returns true if an empty newline should be printer before a node.
func (p *Renderer) spaceBefore(node ast.Node) bool {
	//nolint:gocritic // Written as a switch-case rather than if to allow adding more cases easily.
	switch h := node.(type) {
	case *ast.Heading:
		return h.Level > 1
	}

	return false
}

func (p *Renderer) newline(w io.Writer) {
	p.writtenNewline = true
	_, _ = io.WriteString(w, "\n")
}
