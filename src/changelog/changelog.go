package changelog

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/src/bump"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Version *semver.Version

// Changelog models a machine-readable changelog.
type Changelog struct {
	// Held is true if this changelog should not be released without human intervention.
	Held bool `yaml:"held,omitempty"`
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

	c.Held = c.Held || other.Held
	c.Changes = append(c.Changes, other.Changes...)
	c.Dependencies = append(c.Dependencies, other.Dependencies...)
}

// Empty returns true if this changelog contains no data.
func (c *Changelog) Empty() bool {
	return !c.Held && c.Notes == "" && len(c.Changes) == 0 && len(c.Dependencies) == 0
}

// Entry is a representation of a change that has been made in the code.
type Entry struct {
	// Type holds which bump this change was: A bug fix, a new feature, etc. See EntryType.
	Type EntryType `yaml:"type"`
	// Message is a human-readable one-liner summarizing the change.
	Message string `yaml:"message"`
	// Meta holds information about who made the change and where.
	Meta EntryMeta `yaml:"meta,omitempty"`
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
	Author string `yaml:"author,omitempty"`
	PR     string `yaml:"pr,omitempty"`
	Commit string `yaml:"commit,omitempty"`
}

// Dependency models a dependency that has been changed in the project.
type Dependency struct {
	Name string          `yaml:"name"`
	From *semver.Version `yaml:"from,omitempty"`
	To   *semver.Version `yaml:"to,omitempty"`
	// Link to the changelog for the release of this dependency.
	Changelog string    `yaml:"changelog"`
	Meta      EntryMeta `yaml:"meta,omitempty"`
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
		_, _ = fmt.Fprintf(buf, " from %s", d.From.Original())
	}

	if d.To != nil {
		_, _ = fmt.Fprintf(buf, " to %s", d.To.Original())
	}

	if d.Changelog != "" {
		_, _ = fmt.Fprintf(buf, " [Changelog](%s)", d.Changelog)
	}

	return buf.String()
}

// plainDependency is a helper struct where To and From are strings rather than semver.Version. We use this struct
// to marshal and unmarshal from YAML format because unfortunately, semver.Version does not implement yaml.Marshaler.
type plainDependency struct {
	Name      string    `yaml:"name"`
	From      string    `yaml:"from,omitempty"`
	To        string    `yaml:"to,omitempty"`
	Changelog string    `yaml:"changelog,omitempty"`
	Meta      EntryMeta `yaml:"meta,omitempty"`
}

// MarshalYAML copies the contents of Dependency to a plainDependency and returns it for the generic marshaler to
// encode it.
func (d Dependency) MarshalYAML() (interface{}, error) {
	pd := plainDependency{
		Name:      d.Name,
		Changelog: d.Changelog,
		Meta:      d.Meta,
	}

	if d.To != nil {
		pd.To = d.To.Original()
	}
	if d.From != nil {
		pd.From = d.From.Original()
	}

	return &pd, nil
}

// UnmarshalYAML decodes the node into a plainDependency and copies it over to the real Dependency.
func (d *Dependency) UnmarshalYAML(value *yaml.Node) error {
	pd := plainDependency{}
	err := value.Decode(&pd)
	if err != nil {
		return fmt.Errorf("unmarshalling plain dependency: %w", err)
	}

	d.Name = pd.Name
	d.Changelog = pd.Changelog

	if pd.To != "" {
		d.To, err = semver.NewVersion(pd.To)
		if err != nil {
			return fmt.Errorf("parsing dependency.To %q: %w", pd.To, err)
		}
	}
	if pd.From != "" {
		d.From, err = semver.NewVersion(pd.From)
		if err != nil {
			return fmt.Errorf("parsing dependency.From %q: %w", pd.From, err)
		}
	}

	return nil
}
