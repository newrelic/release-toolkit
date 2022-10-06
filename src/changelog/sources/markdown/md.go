// Package markdown implements sourcing unreleased entries from https://keepachangelog.com/en/1.0.0/
package markdown

import (
	"fmt"
	"io"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/md"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/changelog/sources/markdown/headingdoc"
	log "github.com/sirupsen/logrus"
)

const (
	unreleasedHeader = "unreleased"
	heldHeader       = "held"
)

//nolint:gochecknoglobals
var supportedTypes = []changelog.EntryType{
	changelog.TypeBreaking,
	changelog.TypeSecurity,
	changelog.TypeEnhancement,
	changelog.TypeBugfix,
}

// Markdown is a changelog.Source that produces a changelog.Changelog from a markdown file that resembles the
// [Keep a changelog](https://keepachangelog.com/en/1.0.0/) format.
type Markdown struct {
	reader io.Reader
}

// New returns a new Markdown source that will consume the specified reader.
// No reads are performed until Markdown.Changelog is called.
func New(r io.Reader) Markdown {
	return Markdown{reader: r}
}

// Changelog reads the supplied reader into memory and parses the contents.
// An empty, non-nil changelog will not result in an error, but a very wrongly formatted one will.
func (m Markdown) Changelog() (*changelog.Changelog, error) {
	doc, err := headingdoc.NewFromReader(m.reader)
	if err != nil {
		return nil, fmt.Errorf("parsing markdown: %w", err)
	}

	return builder{doc: doc}.build()
}

// builder is an object which, from a heading doc, can produce a changelog.
type builder struct {
	doc *headingdoc.Doc
	cl  *changelog.Changelog

	visited map[*headingdoc.Doc]bool
}

func (b builder) build() (*changelog.Changelog, error) {
	b.cl = &changelog.Changelog{}

	unreleased := b.doc.FindOne(unreleasedHeader)
	if unreleased == nil {
		log.Debugf("No %q header found, returning empty changelog", unreleasedHeader)
		return b.cl, nil
	}

	b.visited = map[*headingdoc.Doc]bool{}
	log.Tracef("Gathering changelog entries")
	for _, changeType := range supportedTypes {
		log.Debugf("Finding headers for type %q under %q header", changeType, unreleasedHeader)

		changeTypeHeaders := unreleased.Find(string(changeType))
		if len(changeTypeHeaders) == 0 {
			log.Debugf("No header found for type %v", changeType)
			continue
		}

		for _, headerDoc := range changeTypeHeaders {
			b.entriesFromHeader(headerDoc, changeType)
		}
	}

	held := b.doc.FindOne(heldHeader)
	if held != nil {
		b.visited[held] = true

		log.Warnf("Found a L%d %q header, marking changelog as held", held.Level, heldHeader)
		b.cl.Held = true
	}

	log.Tracef("Collecting any other headers under %q as notes", unreleasedHeader)
	b.unvisitedAsNotes(unreleased)

	return b.cl, nil
}

func (b builder) entriesFromHeader(header *headingdoc.Doc, t changelog.EntryType) {
	if len(header.Content) <= 1 {
		log.Warnf("Skipping empty %q header", header.Name)
		return
	}

	b.visited[header] = true

	log.Debugf("Extracting list items under header %q", header.Name)
	// First item of the headingDoc content is always the heading itself, so we skip it for parsing.
	changes := items(header.Content[1:])
	if len(changes) == 0 {
		log.Warnf("No list items found under header %q", header.Name)
	}

	for _, change := range changes {
		b.cl.Changes = append(b.cl.Changes, changelog.Entry{
			Message: change,
			Type:    t,
		})
	}
}

func (b builder) unvisitedAsNotes(header *headingdoc.Doc) {
	notes := strings.Builder{}
	for _, headerDoc := range header.Children {
		if b.visited[headerDoc] {
			continue
		}

		log.Debugf("Adding unvisited header %q as notes", headerDoc.Name)
		for _, node := range headerDoc.Content {
			buf := markdown.Render(node, md.NewRenderer())
			notes.Write(buf)
		}
	}

	b.cl.Notes = notes.String()
}

// items receives a list of ast.Node, and for those nodes which are lists, returns the list items inside.
// Nodes which are not lists are ignored.
func items(content []ast.Node) []string {
	var itemsStr []string

	for _, node := range content {
		list, isList := node.(*ast.List)
		if !isList {
			log.Debugf("Skipping non-list node")
			continue
		}

		for _, listChild := range list.Children {
			item, isItem := listChild.(*ast.ListItem)
			if !isItem {
				log.Warn("Skipping non-list-item child of a list")
				continue
			}

			if len(item.Children) != 1 {
				log.Warn("ListItem has more than 1 children")
				continue
			}

			buf := markdown.Render(item.Children[0], md.NewRenderer())
			itemsStr = append(itemsStr, strings.TrimSpace(string(buf)))
		}
	}

	return itemsStr
}
