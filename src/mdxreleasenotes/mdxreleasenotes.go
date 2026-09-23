package mdxreleasenotes

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
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

// Matches the next `## v...` heading, used to find where the current version's section ends.
var nextVersionHeadingPattern = regexp.MustCompile(`(?m)^##\s+v`)

// Matches a `## v<version> - <YYYY-MM-DD>` heading and captures the release date.
func headingPattern(version string) *regexp.Regexp {
	return regexp.MustCompile(`(?m)^##\s+v` + regexp.QuoteMeta(version) + `\s+-\s+(\d{4}-\d{2}-\d{2})\s*$`)
}

// Matches a `### ...` changelog subsection heading and captures its title.
var subsectionHeadingPattern = regexp.MustCompile(`(?m)^###\s+(.*)$`)

// Render extracts the section for version from changelog and writes the rendered MDX to w.
func (r Renderer) Render(w io.Writer, changelog string, version string) error {
	changelog = strings.ReplaceAll(changelog, "\r", "")

	loc := headingPattern(version).FindStringSubmatchIndex(changelog)
	if loc == nil {
		return fmt.Errorf("%w: version %q", ErrHeadingNotFound, version)
	}

	releaseDate := changelog[loc[2]:loc[3]]

	bodyStart := loc[1]

	bodyEnd := len(changelog)
	if nextLoc := nextVersionHeadingPattern.FindStringIndex(changelog[bodyStart:]); nextLoc != nil {
		bodyEnd = bodyStart + nextLoc[0]
	}

	body := changelog[bodyStart:bodyEnd]

	bulletsBySection := make(map[string][]string, len(r.Sections))
	for _, section := range r.Sections {
		bulletsBySection[section.Key] = extractBullets(body, section.Keywords)
	}

	if err := r.renderFrontmatter(w, version, releaseDate, bulletsBySection); err != nil {
		return err
	}

	if err := r.renderBody(w, version, bulletsBySection); err != nil {
		return err
	}

	return nil
}

func extractBullets(body string, keywords []string) []string {
	headings := subsectionHeadingPattern.FindAllStringSubmatchIndex(body, -1)

	for i, headingMatch := range headings {
		heading := body[headingMatch[2]:headingMatch[3]]
		if !containsAny(heading, keywords) {
			continue
		}

		contentStart := headingMatch[1]
		contentEnd := len(body)
		if i+1 < len(headings) {
			contentEnd = headings[i+1][0]
		}

		return cleanBullets(body[contentStart:contentEnd])
	}

	return nil
}

func containsAny(haystack string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(haystack, needle) {
			return true
		}
	}

	return false
}

func cleanBullets(content string) []string {
	var items []string

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimSpace(line)

		if line != "" {
			items = append(items, line)
		}
	}

	return items
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
