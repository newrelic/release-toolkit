package pretty

import (
	"io"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
)

// Renderer is a markdown.Renderer wrapper that adds empty newlines after and before certain elements.
type Renderer struct {
	markdown.Renderer
}

// New returns a new Renderer.
// TODO: It would be great to be able to configure this to have customizable "before" and "after" spacing.
func New(inner markdown.Renderer) Renderer {
	return Renderer{
		Renderer: inner,
	}
}

// RenderNode intercepts the inner renderer's RenderNode, implementing the newline logic.
func (p Renderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	switch {
	case entering && p.spaceBefore(node):
		fallthrough
	case !entering && p.spaceAfter(node):
		_, _ = io.WriteString(w, "\n")
	}

	return p.Renderer.RenderNode(w, node, entering)
}

// spaceAfter returns true if an empty newline should be printer after a node.
func (p Renderer) spaceAfter(node ast.Node) bool {
	// TODO: This seemed like a great idea at first, but it does not behave as expected if we have a list and a heading
	// just after it, as it will add a newline for each resulting in two empty lines.
	// For now, we will just leave with paragraphs cuddled with lists.

	//nolint:gocritic // This is commented-out code and therefore it does not need space before text.
	//switch node.(type) {
	//case *ast.List:
	//	return true
	//}

	return false
}

// spaceBefore returns true if an empty newline should be printer before a node.
func (p Renderer) spaceBefore(node ast.Node) bool {
	//nolint:gocritic // Written as a switch-case rather than if to allow adding more cases easily.
	switch h := node.(type) {
	case *ast.Heading:
		return h.Level > 1
	}

	return false
}
