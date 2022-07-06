package headingdoc

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gomarkdown/markdown/ast"
	log "github.com/sirupsen/logrus"
)

// Doc is a convenience type that sorts a markdown document by headers.
type Doc struct {
	// Name is the name of the header of this section
	Name string
	// Level is the level of the header
	Level int
	// Content is a list of the markdown nodes that came after this header, and before any other header
	Content []ast.Node
	// Children is a list of headers that came after this header and are a lower level than this header
	Children []*Doc
	// Parent is the higher-level header that appeared in the document before this one.
	Parent *Doc
}

var (
	ErrEmptyName   = errors.New("heading has no name")
	ErrNotL1       = errors.New("first header is not level 1")
	ErrNotDocument = errors.New("markdown ast node is not a document node")
)

// New takes a root (ast.Document) markdown node and buidls a Doc with it.
func New(root ast.Node) (*Doc, error) {
	current := &Doc{}
	docRoot := current
	var err error

	doc, isDoc := root.(*ast.Document)
	if !isDoc {
		return nil, ErrNotDocument
	}

	for _, child := range doc.Children {
		current, err = current.append(child)
		if err != nil {
			return nil, fmt.Errorf("parsing markdown headers: %w", err)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("adding node to doc: %w", err)
	}

	return docRoot, nil
}

// Find searches this header and its subheaders for one matching the given name and level.
// Name is compared loosely, by checking in a case insensitive way whether the header contains the specified name.
func (hd *Doc) Find(name string) []*Doc {
	if strings.Contains(strings.ToLower(hd.Name), strings.ToLower(name)) {
		log.Tracef("Header %q matches %q, returning itself", hd.Name, name)
		return []*Doc{hd}
	}

	log.Tracef("Header %q does not match %q, recursing through children", hd.Name, name)
	var found []*Doc
	for _, child := range hd.Children {
		found = append(found, child.Find(name)...)
	}

	return found
}

// FindOne returns the first result that Find would return.
func (hd *Doc) FindOne(name string) *Doc {
	list := hd.Find(name)
	if len(list) == 0 {
		return nil
	}

	return list[0]
}

// append is used to populate the Doc from markdown nodes.
func (hd *Doc) append(node ast.Node) (*Doc, error) {
	heading, isHeading := node.(*ast.Heading)
	if !isHeading {
		if hd.Name == "" {
			log.Trace("Skipping content before first header")
			return hd, nil
		}

		log.Trace("Adding non-heading content to heading")
		hd.Content = append(hd.Content, node)

		return hd, nil
	}

	if hd.Name == "" {
		log.Trace("Attempted to append heading to empty node, initializing")

		name := headingName(heading)
		if name == "" {
			return nil, ErrEmptyName
		}

		if heading.Level != 1 && hd.Parent == nil {
			return nil, ErrNotL1
		}

		hd.Name = name
		hd.Level = heading.Level

		// Append header node to content. We need this to reconstruct the MD document without much effort.
		hd.Content = append(hd.Content, node)

		return hd, nil
	}

	if heading.Level <= hd.Level {
		if hd.Parent == nil {
			log.Warnf("Header is a sibling of the top-level header, which is not supported. Dropping header")
			return hd, nil
		}

		log.Trace("Header to be appended is the same or greater level as me, forwarding request to my parent")
		return hd.Parent.append(node)
	}

	child := &Doc{
		Parent: hd,
	}

	hd.Children = append(hd.Children, child)

	// Initialize child, which will return itself.
	return child.append(heading)
}

// headingName is a helper function that returns the name of an ast.Heading, by getting the text of the first child.
func headingName(h *ast.Heading) string {
	if len(h.Children) != 1 {
		return ""
	}

	leaf := h.Children[0].AsLeaf()
	if leaf == nil {
		return ""
	}

	return string(leaf.Literal)
}
