// Package merger contains the Merger object, which incorporates a changelog.Changelog
// into a full markdown changelog.
package merger

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/changelog"
	"github.com/newrelic/release-toolkit/changelog/renderer"
	log "github.com/sirupsen/logrus"
)

// Merger is an object that can incorporate a changelog.Changelog (section) into an existing CHANGELOG.md document.
type Merger struct {
	// ReleasedOn is a function that returns the date in which the new section was released. It defaults to time.Now.
	ReleasedOn func() time.Time

	// version holds the in which the new changelog was released.
	version *semver.Version
	// ch is the already populated, partial changelog containing the latest changes.
	ch *changelog.Changelog
}

// New creates a Merger that will integrate the supplied changelog.Changelog into a full Markdown document that contains
// changelogs for older versions. The inserted section will be marked as being the changelog for the version specified
// in the supplied semver.Version, and also marked as released on the date returned by Merger.ReleasedOn.
// Merger is an immutable object and does not modify the changelog.Changelog object, nor the original document supplied
// to Merge.
func New(ch *changelog.Changelog, newVersion *semver.Version) Merger {
	return Merger{
		ReleasedOn: time.Now,
		ch:         ch,
		version:    newVersion,
	}
}

var (
	unreleasedHeader = regexp.MustCompile(`(?i)^##\s*unreleased`)
	heldHeader       = regexp.MustCompile(`(?i)^##\s*held`)
	l2Header         = regexp.MustCompile(`^##\s*\w`)
)

// Merge uses the configured changelog and version to read the current, full changelog in Markdown format from
// srcChangelog, and write to dst a new full changelog containing the entries specified in the changelog.Changelog
// object that was passed to New.
// The source file is left intact and the changelog.Changelog object supplied to New are not modified.
func (m Merger) Merge(srcChangelog io.Reader, dst io.Writer) error {
	newSection := &bytes.Buffer{}

	rdr := renderer.New(m.ch)
	rdr.Next = m.version
	rdr.ReleasedOn = m.ReleasedOn

	err := rdr.Render(newSection)
	if err != nil {
		return fmt.Errorf("rendering new changelog: %w", err)
	}
	// We add two newlines to the new section to have it properly formatted in the rest of the doc.
	newSection.WriteString("\n\n")

	scanner := bufio.NewScanner(srcChangelog)

	// Ignore acts like a sort-of state-machine that tells the consuming loop whether the subsequent line needs to be
	// copied over or ignored.
	// Ignore is set to true when a specific L2 header we want to ignore is found, and set back to false when any other
	// L2 header is found.
	ignore := false

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case unreleasedHeader.MatchString(line):
			log.Tracef("Unreleased header found, printing header and ignoring section")
			_, _ = fmt.Fprintf(dst, "%s\n\n", line)
			ignore = true
		case heldHeader.MatchString(line):
			log.Tracef("Held header found, ignoring both header and section")
			ignore = true
		case l2Header.MatchString(line):
			log.Tracef("L2 header found, including section")
			ignore = false

			// Copy the new section now before writing this header.
			// Calls to this function on subsequent L2 headers are noop as the newSection buffer is already consumed.
			_, err = io.Copy(dst, newSection)
			if err != nil {
				return fmt.Errorf("inserting new changelog: %w", err)
			}
		}

		if !ignore {
			_, err = fmt.Fprintln(dst, line)
			if err != nil {
				return fmt.Errorf("copying line: %w", err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading header: %w", err)
	}

	return nil
}
