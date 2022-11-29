package markdown

import (
	"io"
	"reflect"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
)

// FilterRenderer wraps a markdown.Renderer skips calling it for certain node types.
// This allows to use the default markdown renderer with types it does not support without it panicking.
type FilterRenderer struct {
	filtered map[reflect.Type]bool
	markdown.Renderer
}

// NewFilterRenderer returns a FilterRenderer given an underlying renderer and a list of nodes whose type will be ignored.
func NewFilterRenderer(inner markdown.Renderer, ignore ...ast.Node) FilterRenderer {
	filtered := map[reflect.Type]bool{}
	for _, iface := range ignore {
		filtered[reflect.TypeOf(iface)] = true
	}

	return FilterRenderer{
		filtered: filtered,
		Renderer: inner,
	}
}

// RenderNode implements markdown.Renderer with the additional filtering logic.
func (fr FilterRenderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	if fr.filtered[reflect.TypeOf(node)] {
		return ast.GoToNext
	}

	return fr.Renderer.RenderNode(w, node, entering)
}
