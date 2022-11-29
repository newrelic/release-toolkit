package markdown_test

import (
	"io"
	"testing"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/md"
	"github.com/gomarkdown/markdown/parser"
	filterer "github.com/newrelic/release-toolkit/src/changelog/sources/markdown"
)

const list = `
- This is a list
- This line has trailing spaces   
- This one has not
`

func TestRendererPanics(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			t.Log("Markdown renderer panicked as expected, FilterRenderer is still necessary.")
		}
	}()

	doc := parser.New().Parse([]byte(list))
	markdown.Render(doc, md.NewRenderer())

	t.Errorf("Makrdown renderer did not panic when rendering an *ast.Hardbreak. Our homemade FilterRenderer should be removed.")
}

type mockRenderer struct {
	called int
}

func (mr *mockRenderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	mr.called++
	return ast.GoToNext
}
func (mr *mockRenderer) RenderHeader(w io.Writer, ast ast.Node) {}
func (mr *mockRenderer) RenderFooter(w io.Writer, ast ast.Node) {}

func TestFilterRenderer(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		nodes        []ast.Node
		ignore       []ast.Node
		expectCalled int
	}{
		{
			name: "Skips_Nodes",
			nodes: []ast.Node{
				&ast.Paragraph{},
				&ast.Heading{},
				&ast.Image{},
				&ast.List{},
				&ast.ListItem{},
			},
			ignore: []ast.Node{
				&ast.Image{},
				&ast.List{},
			},
			expectCalled: 3,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockRenderer{}
			fr := filterer.NewFilterRenderer(mock, tc.ignore...)

			for _, node := range tc.nodes {
				fr.RenderNode(nil, node, false)
			}

			if mock.called != tc.expectCalled {
				t.Fatalf("Expected to render %d nodes, rendered %d", tc.expectCalled, mock.called)
			}
		})
	}
}
