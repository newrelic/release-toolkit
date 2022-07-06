// Package merger contains the Merger object, which incorporates a changelog.Changelog
// into a full markdown changelog.
package merger

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/changelog"
	"github.com/newrelic/release-toolkit/changelog/renderer"
	"github.com/newrelic/release-toolkit/changelog/sources/markdown/headingdoc"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

const (
	unreleasedHeader = "unreleased"
)

var (
	ErrNoUnreleased = errors.New("unreleased header not found")
)

type Merger struct {
	ReleasedOn func() time.Time

	version *semver.Version
	ch      *changelog.Changelog
}

func New(ch *changelog.Changelog, newVersion *semver.Version) Merger {
	return Merger{
		ReleasedOn: time.Now,
		ch:         ch,
		version:    newVersion,
	}
}

func (m Merger) Merge(srcChangelog io.Reader, dst io.Writer) error {
	fullDoc, err := headingdoc.NewFromReader(srcChangelog)
	if err != nil {
		return fmt.Errorf("parsing original changelog: %w", err)
	}

	mdChangelog := &bytes.Buffer{}
	rdr := renderer.New(m.ch)
	rdr.Next = m.version
	rdr.ReleasedOn = m.ReleasedOn
	err = rdr.Render(mdChangelog)
	if err != nil {
		return fmt.Errorf("rendering new changelog: %w", err)
	}

	newDoc, err := headingdoc.NewFromReader(mdChangelog)
	if err != nil {
		return fmt.Errorf("parsing rendered changelog: %w", err)
	}

	log.Tracef("Emptying out unreleased header")
	ur := fullDoc.FindOne(unreleasedHeader)
	if ur == nil {
		return ErrNoUnreleased
	}

	ur.Children = nil
	ur.Content = ur.Content[:1]

	log.Tracef("Appending new version header")
	unreleasedIndex := slices.Index(fullDoc.Children, ur)
	if unreleasedIndex == -1 {
		return ErrNoUnreleased
	}

	// Create a new list of children with room for the unreleased header.
	newChildren := make([]*headingdoc.Doc, unreleasedIndex+1)
	// Copy over the list of children of the original document, stopping after the unreleased header.
	copy(newChildren, fullDoc.Children)
	// Insert the new section.
	newChildren = append(newChildren, newDoc)
	// Finally, append the rest of (older) children.
	newChildren = append(newChildren, fullDoc.Children[unreleasedIndex+1:]...)

	fullDoc.Children = newChildren

	err = fullDoc.Render(dst)
	if err != nil {
		return fmt.Errorf("rendering merged doc: %w", err)
	}

	return nil
}
