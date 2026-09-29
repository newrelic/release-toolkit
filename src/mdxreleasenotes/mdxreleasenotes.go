package mdxreleasenotes

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/newrelic/release-toolkit/src/changelog/sources/markdown"
	"github.com/newrelic/release-toolkit/src/changelog/sources/markdown/headingdoc"
)

var ErrHeadingNotFound = errors.New("could not find heading in changelog")

// Section identifies a group of changelog bullets by the docs-site array they should be collected into, and the
// heading keywords that mark a `### ...` changelog subsection as belonging to that group.
type Section struct {
	// YAML frontmatter array name.
	Key string
	// Heading used in the rendered markdown body.
	Title string
	// Keywords are matched against `### ...` heading text to decide whether a changelog subsection belongs to
	// this Section.
	Keywords []string
}

// Renderer extracts a version's section from a changelog and renders it as MDX frontmatter plus a markdown body.
type Renderer struct {
	// Name used in the `subject` frontmatter field.
	Subject string
	// `owner/name` repository used to build the release notes link.
	Repo     string
	Sections []Section
}

// DefaultSections mirrors the sections used by New Relic's changelog conventions: Enhancements/Features,
// Bug fixes, and Security notices.
//
//nolint:gochecknoglobals // This is a read-only default configuration value, akin to a constant.
var DefaultSections = []Section{
	{Key: "features", Title: "New features", Keywords: []string{"Enhancements", "Features"}},
	// The release-toolkit doesn't have an "Enhancementes" heading. It is kept as a
	// frontmatter field to comply with NewRelic public docs site schema.
	// https://github.com/newrelic/docs-website/blob/develop/templates/release-notes-template.mdx
	{Key: "enhancements", Title: "Improvements and enhancements", Keywords: nil},
	{Key: "bugs", Title: "Bug fixes", Keywords: []string{"Bug fixes"}},
	{Key: "security", Title: "Security updates", Keywords: []string{"Security notices"}},
}

// Render extracts the section for version from changelog and writes the rendered MDX to w.
func (r Renderer) Render(w io.Writer, changelog string, version string) error {
	doc, err := headingdoc.NewFromReader(strings.NewReader(changelog))
	if err != nil {
		return fmt.Errorf("parsing changelog: %w", err)
	}

	versionDoc := doc.FindOne("v" + version)
	if versionDoc == nil {
		return fmt.Errorf("%w: version %q", ErrHeadingNotFound, version)
	}

	_, releaseDate, _ := strings.Cut(versionDoc.Name, "-")
	releaseDate = strings.TrimSpace(releaseDate)

	bulletsBySection := make(map[string][]string, len(r.Sections))
	for _, section := range r.Sections {
		bulletsBySection[section.Key] = extractBullets(versionDoc, section.Keywords)
	}

	if err := r.renderFrontmatter(w, version, releaseDate, bulletsBySection); err != nil {
		return err
	}

	if err := r.renderBody(w, version, bulletsBySection); err != nil {
		return err
	}

	return nil
}

// extractBullets returns the bullet items under the subsection of versionDoc matching one of keywords.
func extractBullets(versionDoc *headingdoc.Doc, keywords []string) []string {
	for _, keyword := range keywords {
		sub := versionDoc.FindOne(keyword)
		if sub == nil {
			continue
		}

		// First item of a Doc's content is always its own heading, so we skip it when extracting items.
		return markdown.Items(sub.Content[1:])
	}

	return nil
}

func (r Renderer) renderFrontmatter(w io.Writer, version, releaseDate string, bulletsBySection map[string][]string) error {
	buf := &strings.Builder{}

	fmt.Fprintln(buf, "---")
	fmt.Fprintf(buf, "subject: %s\n", r.Subject)
	fmt.Fprintf(buf, "releaseDate: '%s'\n", releaseDate)
	fmt.Fprintf(buf, "version: %s\n", version)

	for _, section := range r.Sections {
		fmt.Fprintf(buf, "%s: %s\n", section.Key, yamlList(bulletsBySection[section.Key]))
	}

	fmt.Fprintln(buf, "---")
	fmt.Fprintln(buf)

	_, err := io.WriteString(w, buf.String())
	if err != nil {
		return fmt.Errorf("writing frontmatter: %w", err)
	}

	return nil
}

func yamlList(items []string) string {
	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = "'" + strings.ReplaceAll(item, "'", "''") + "'"
	}

	return "[" + strings.Join(quoted, ", ") + "]"
}

func (r Renderer) renderBody(w io.Writer, version string, bulletsBySection map[string][]string) error {
	buf := &strings.Builder{}

	for _, section := range r.Sections {
		bullets := bulletsBySection[section.Key]
		if len(bullets) == 0 {
			continue
		}

		fmt.Fprintf(buf, "### %s\n\n", section.Title)
		for _, bullet := range bullets {
			fmt.Fprintf(buf, "- %s\n", bullet)
		}
		fmt.Fprintln(buf)
	}

	releaseNotesLink := fmt.Sprintf("https://github.com/%s/releases/tag/%s", r.Repo, version)
	fmt.Fprintf(buf, "For a detailed description of changes, see the [release notes](%s).\n", releaseNotesLink)

	_, err := io.WriteString(w, buf.String())
	if err != nil {
		return fmt.Errorf("writing body: %w", err)
	}

	return nil
}
