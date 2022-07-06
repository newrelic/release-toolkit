package renderer

import (
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/changelog"
)

// Stringer is anything that can be printed as a list entry on the changelog. changelog.Dependency and changelog.Entry
// implement this interface.
type Stringer interface {
	String() string
}

type Renderer struct {
	ReleasedOn func() time.Time
	Next       *semver.Version
	changelog  *changelog.Changelog
}

func New(c *changelog.Changelog) Renderer {
	return Renderer{
		changelog:  c,
		ReleasedOn: time.Now,
		Next:       semver.MustParse("v0.0.0"),
	}
}

type parsedChangelog struct {
	Version  string
	Date     string
	Notes    string
	Sections map[string][]Stringer
}

func (r Renderer) Render(w io.Writer) error {
	parsed := r.parse()
	tpl, err := template.New("changelog").Parse(strings.TrimSpace(markdownTemplate))
	if err != nil {
		return fmt.Errorf("internal, parsing template: %w", err)
	}

	err = tpl.Execute(w, parsed)
	if err != nil {
		return fmt.Errorf("populating template: %w", err)
	}

	return nil
}

func (r Renderer) parse() parsedChangelog {
	parsed := parsedChangelog{
		Version:  "v" + r.Next.String(),
		Date:     r.ReleasedOn().Format("2006-01-02"),
		Notes:    strings.TrimSpace(r.changelog.Notes),
		Sections: map[string][]Stringer{},
	}

	for _, entry := range r.changelog.Changes {
		parsed.Sections[string(entry.Type)] = append(parsed.Sections[string(entry.Type)], entry)
	}

	for _, dep := range r.changelog.Dependencies {
		parsed.Sections[string(changelog.TypeDependency)] = append(parsed.Sections[string(changelog.TypeDependency)], dep)
	}

	return parsed
}
