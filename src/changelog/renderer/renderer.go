package renderer

import (
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/src/changelog"
)

// Stringer is anything that can be printed as a list entry on the changelog. changelog.Dependency and changelog.Entry
// implement this interface.
type Stringer interface {
	String() string
}

// Renderer outputs a human-readable, markdown version of a changelog.
type Renderer struct {
	// If non-nil, the output will include a level 2 header with the version to which this changelog corresponds.
	Next *semver.Version
	// If non-nil and Next is non-nil, the level 2 header including the version will also include the date returned by
	// this function, signifying that the version to which this changelog corresponds was released on said date.
	ReleasedOn func() time.Time

	changelog               *changelog.Changelog
	deduplicateDependencies bool
}

type OptionFunc func(r *Renderer)

func WithDeduplicateDependencies() OptionFunc {
	return func(r *Renderer) {
		r.deduplicateDependencies = true
	}
}

func New(c *changelog.Changelog, opts ...OptionFunc) Renderer {
	r := &Renderer{
		changelog: c,
	}
	for _, opt := range opts {
		opt(r)
	}

	return *r
}

type parsedChangelog struct {
	Version  string
	Date     string
	Notes    string
	Sections map[string][]Stringer
}

// Render writes the markdown representation of a changelog to the specified writer.
func (r Renderer) Render(w io.Writer) error {
	parsed := r.parse()
	tpl, err := template.New("changelog").Parse(strings.TrimSpace(markdownTemplate))
	if err != nil {
		return fmt.Errorf("internal, parsing template: %w", err)
	}

	// For templates to be sane and readable, we need to put spaces between sections _after_ them. This comes with the
	// problem of the last section of the doc also printing those spaces, leading to two empty newlines.
	// As we need to chomp those newlines, we must write to an intermediate buffer, trim it, and then copy back to the
	// supplied writer.
	buf := &strings.Builder{}
	err = tpl.Execute(buf, parsed)
	if err != nil {
		return fmt.Errorf("populating template: %w", err)
	}

	_, err = fmt.Fprint(w, strings.TrimSpace(buf.String()))
	if err != nil {
		return fmt.Errorf("writing output to writer: %w", err)
	}

	return nil
}

func (r Renderer) parse() parsedChangelog {
	parsed := parsedChangelog{
		Notes:    strings.TrimSpace(r.changelog.Notes),
		Sections: map[string][]Stringer{},
	}

	if r.Next != nil {
		parsed.Version = "v" + r.Next.String()
	}

	if r.ReleasedOn != nil {
		parsed.Date = r.ReleasedOn().Format("2006-01-02")
	}

	for _, entry := range r.changelog.Changes {
		parsed.Sections[string(entry.Type)] = append(parsed.Sections[string(entry.Type)], entry)
	}

	deps := r.changelog.Dependencies
	if r.deduplicateDependencies {
		deps = deduplicateDependencies(deps)
	}

	for _, dep := range deps {
		parsed.Sections[string(changelog.TypeDependency)] = append(parsed.Sections[string(changelog.TypeDependency)], dep)
	}

	return parsed
}

// Dependencies are sorted in ascending order. We keep the latest that should be the one with the latest semVer.
func deduplicateDependencies(dependencies []changelog.Dependency) []changelog.Dependency {
	dedupDeps := []changelog.Dependency{}
	for _, dep := range dependencies {
		found := false
		for i := range dedupDeps {
			if dedupDeps[i].Name == dep.Name {
				found = true
				dedupDeps[i] = dep
			}
		}
		if !found {
			dedupDeps = append(dedupDeps, dep)
		}
	}
	return dedupDeps
}
