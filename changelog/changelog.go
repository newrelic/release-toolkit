package changelog

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/bump"
	log "github.com/sirupsen/logrus"
)

// Changelog models a machine-readable changelog.
type Changelog struct {
	// Held is true if this changelog should not be released without human intervention.
	Held bool
	// Notes is a markdown snippet that will be rendered just below the version header.
	Notes string `yaml:"notes"`
	// Changes is a list of changes that have been made since the last release.
	// They have a description and a bump.Type, and may optionally include metadata about the author, PR, or commit when
	// they were introduced.
	Changes []Entry `yaml:"changes"`
	// Dependencies is a list of the dependencies that have been bumped, optionally including info about from which
	// to which version they were.
	Dependencies []Dependency `yaml:"dependencies"`
}

// Merge appends changes noted in other changelog to the current one.
func (c *Changelog) Merge(other *Changelog) {
	if other.Notes != "" {
		// TODO: Notes is a markdown section. For certain use cases it might be better to perform a "markdown merge",
		// so contents of sections with the same name are merged rather than sections getting duplicated.
		// For now, the naive implementation should be enough.
		if c.Notes != "" {
			log.Warn("Naively merging changelog notes, output might not be ideal")
		}
		c.Notes = strings.TrimSpace(strings.TrimSpace(c.Notes) + "\n\n" + other.Notes)
	}

	c.Changes = append(c.Changes, other.Changes...)
	c.Dependencies = append(c.Dependencies, other.Dependencies...)
}

// Entry is a representation of a change that has been made in the code.
type Entry struct {
	// Message is a human-readable one-liner summarizing the change.
	Message string
	// Type holds which bump this change was: A bug fix, a new feature, etc. See EntryType.
	Type EntryType
	// Meta holds information about who made the change and where.
	Meta EntryMeta
}

// EntryType encodes the nature of the change.
type EntryType string

const (
	TypeEnhancement = EntryType("enhancement")
	TypeBugfix      = EntryType("bugfix")
	TypeSecurity    = EntryType("security")
	TypeBreaking    = EntryType("breaking")
	// TypeDependency is the entry bump for a dependency bump. It is better, however, to encode dependency changes in
	// Changelog.Dependencies rather than Changelog.Changes as that allows for smarter semver bumping and richer format.
	TypeDependency = EntryType("dependency")
)

// BumpType returns which version should be bumped due to this change.
func (e Entry) BumpType() bump.Type {
	//nolint:exhaustive
	switch e.Type {
	case TypeBugfix:
		return bump.Patch
	case TypeEnhancement:
		return bump.Minor
	case TypeSecurity:
		return bump.Minor
	case TypeBreaking:
		return bump.Major
	}

	return bump.None
}

// Strings outputs a human-readable one-liner of the change, including meta information if found.
func (e Entry) String() string {
	buf := &strings.Builder{}
	buf.WriteString(e.Message)

	if e.Meta.Author != "" {
		_, _ = fmt.Fprintf(buf, ", by %s", e.Meta.Author)
	}

	if e.Meta.PR != "" {
		_, _ = fmt.Fprintf(buf, " (%s)", e.Meta.PR)
	} else if e.Meta.Commit != "" {
		_, _ = fmt.Fprintf(buf, " (%s)", e.Meta.Commit)
	}

	return buf.String()
}

// EntryMeta holds information about who made the change and where.
type EntryMeta struct {
	Author string
	PR     string
	Commit string
}

// Dependency models a dependency that has been changed in the project.
type Dependency struct {
	Name string
	From *semver.Version
	To   *semver.Version
}

// BumpType returns which version should be bumped due to this dependency update.
// In practice, this is the same as the bump the dependency had.
func (d Dependency) BumpType() bump.Type {
	if d.From == nil || d.To == nil {
		log.Debugf("Dependency %s has unknown to/from versions, assuming patch bump", d.Name)
		return bump.Patch
	}

	return bump.From(d.From, d.To)
}

func (d Dependency) Change() string {
	if d.From == nil || d.To == nil {
		return "Updated"
	}

	switch {
	case d.To.LessThan(d.From):
		return "Downgraded"
	case d.To.GreaterThan(d.From):
		return "Upgraded"
	default:
		return "Updated"
	}
}

// Strings outputs a human-readable one-liner of the dependency update, including extra information if found.
func (d Dependency) String() string {
	buf := &strings.Builder{}

	buf.WriteString(d.Change())

	_, _ = fmt.Fprintf(buf, " %s", d.Name)

	if d.From != nil {
		_, _ = fmt.Fprintf(buf, " from v%s", d.From.String())
	}

	if d.To != nil {
		_, _ = fmt.Fprintf(buf, " to v%s", d.To.String())
	}

	return buf.String()
}
