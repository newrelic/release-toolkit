package pretty_test

import (
	"strings"
	"testing"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/md"
	"github.com/gomarkdown/markdown/parser"
	"github.com/google/go-cmp/cmp"
	"github.com/newrelic/release-toolkit/changelog/sources/markdown/pretty"
)

func TestRenderer_RenderNode(t *testing.T) {
	orig := `# I'm a markdown document
Here's a pragraph
## I don't have much spaces
As you can see
### And that makes me
## Incredibly unreadable
## Dontyathink
- Heck yes
- Here's another

I think so
` + "```" + `
Here, have a code block
` + "```" + `
`

	// Lack of newline between list and paragraph is, unfortunately, expected.
	expected := `# I'm a markdown document
Here's a pragraph

## I don't have much spaces
As you can see

### And that makes me

## Incredibly unreadable

## Dontyathink
- Heck yes
- Here's another

I think so

` + "```" + `
Here, have a code block
` + "```" + `
`

	root := markdown.Parse([]byte(orig), parser.New())
	buf := &strings.Builder{}
	buf.Write(markdown.Render(root, pretty.New(md.NewRenderer())))

	actual := buf.String()
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("Output markdown is not as expected:\n%s", diff)
	}
}
