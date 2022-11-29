package markdown_test

import (
	"testing"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/md"
	"github.com/gomarkdown/markdown/parser"
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
