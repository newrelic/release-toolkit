// Package mdxreleasenotes extracts a single version's section from a markdown CHANGELOG.md and renders it as an
// MDX file suitable for publishing on a docs site.
package mdxreleasenotes

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Section identifies a group of changelog bullets by the docs-site array they should be collected into, and the
// heading keywords (case-insensitive) that mark a `### ...` changelog subsection as belonging to that group.
type Section struct {
	// Key is the YAML frontmatter array name and the plural noun used in the rendered body heading.
	Key string
	// Title is the heading used in the rendered markdown body, e.g. "New features".
	Title string
	// Keywords are matched case-insensitively against `### ...` heading text to decide whether a changelog
	// subsection belongs to this Section.
	Keywords []string
}

// Renderer extracts a version's section from a changelog and renders it as MDX frontmatter plus a markdown body.
type Renderer struct {
	// Subject is the product/component name stamped into the `subject` frontmatter field.
	Subject string
	// Repo is the `owner/name` GitHub repository used to build the release notes link, e.g.
	// "newrelic/newrelic-agent-control".
	Repo string
	// Sections defines, in order, which changelog subsections are collected and how they are rendered. The order
	// here is also the order sections appear in the frontmatter and in the rendered body.
	Sections []Section
}

// DefaultSections mirrors the sections used by New Relic's changelog conventions: Enhancements/Features,
// Bug fixes, and Security notices.
//
//nolint:gochecknoglobals // This is a read-only default configuration value, akin to a constant.
var DefaultSections = []Section{
	{Key: "features", Title: "New features", Keywords: []string{"Enhancements", "Features"}},
	// Agent Control's changelog has no separate "new feature" bucket — its "Enhancements" heading is matched by
	// features above, so enhancements has no keywords of its own and always renders empty. It is kept as a
	// frontmatter field for parity with the docs-site schema shared across subjects.
	{Key: "enhancements", Title: "Improvements and enhancements", Keywords: nil},
	{Key: "bugs", Title: "Bug fixes", Keywords: []string{"Bug fixes", "Fixes", "Bugfixes"}},
	{Key: "security", Title: "Security updates", Keywords: []string{"Security notices", "Security"}},
}

var prRefPattern = regexp.MustCompile(`\s*\(#\d+\)\s*$`)

var commitRefPattern = regexp.MustCompile(`\s*\([0-9a-fA-F]{7,40}\)\s*$`)

// headingPattern matches a `## v<version>` heading, optionally followed by `- <YYYY-MM-DD>`, and captures the date.
// The body following the heading is sliced separately, since Go's RE2 engine does not support lookahead.
func headingPattern(version string) *regexp.Regexp {
	pattern := `^##\s+v` + regexp.QuoteMeta(version) + `(?:\s+-\s+(\d{4}-\d{2}-\d{2}))?\s*$`

	return regexp.MustCompile(`(?m)` + pattern)
}

// nextVersionHeadingPattern matches the next `## v...` heading, used to find where the current version's section
// ends.
var nextVersionHeadingPattern = regexp.MustCompile(`(?m)^##\s+v`)

// Render extracts the section for version from changelog and writes the rendered MDX to w.
func (r Renderer) Render(w io.Writer, changelog string, version string) error {
	changelog = strings.ReplaceAll(changelog, "\r", "")

	loc := headingPattern(version).FindStringSubmatchIndex(changelog)
	if loc == nil {
		return fmt.Errorf("could not find heading %q in changelog", "## v"+version)
	}

	releaseDate := ""
	if loc[2] != -1 {
		releaseDate = changelog[loc[2]:loc[3]]
	}
	if releaseDate == "" {
		return fmt.Errorf("heading %q has no '- <YYYY-MM-DD>' date", "## v"+version)
	}

	bodyStart := loc[1]
	if bodyStart < len(changelog) && changelog[bodyStart] == '\n' {
		bodyStart++
	}

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

	fmt.Fprintf(buf, "For a detailed description of changes, see the [release notes](%s).\n", releaseURL(r.Repo, version))

	_, err := io.WriteString(w, buf.String())
	if err != nil {
		return fmt.Errorf("writing body: %w", err)
	}

	return nil
}

func releaseURL(repo, version string) string {
	return fmt.Sprintf("https://github.com/%s/releases/tag/%s", repo, version)
}

// subsectionHeadingPattern matches a `### ...` changelog subsection heading and captures its title.
var subsectionHeadingPattern = regexp.MustCompile(`(?m)^###\s+(.*)$`)

// extractBullets returns the cleaned bullet items from the first `### ...` subsection of body whose heading
// contains one of keywords (case-insensitive, symbols such as emoji are ignored by substring matching).
func extractBullets(body string, keywords []string) []string {
	headings := subsectionHeadingPattern.FindAllStringSubmatchIndex(body, -1)

	for i, headingMatch := range headings {
		heading := body[headingMatch[2]:headingMatch[3]]
		if !containsAnyFold(heading, keywords) {
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

func containsAnyFold(haystack string, needles []string) bool {
	haystack = strings.ToLower(haystack)
	for _, needle := range needles {
		if strings.Contains(haystack, strings.ToLower(needle)) {
			return true
		}
	}

	return false
}

func cleanBullets(content string) []string {
	var items []string

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)

		if !strings.HasPrefix(line, "* ") && !strings.HasPrefix(line, "- ") {
			continue
		}

		item := strings.TrimSpace(line[2:])
		item = prRefPattern.ReplaceAllString(item, "")
		item = commitRefPattern.ReplaceAllString(item, "")
		item = strings.TrimSpace(item)

		if item != "" {
			items = append(items, item)
		}
	}

	return items
}

// yamlList renders items as a flow-style single-quoted YAML array, escaping embedded single quotes by doubling
// them, per YAML single-quoted scalar rules.
func yamlList(items []string) string {
	if len(items) == 0 {
		return "[]"
	}

	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = "'" + strings.ReplaceAll(item, "'", "''") + "'"
	}

	return "[" + strings.Join(quoted, ", ") + "]"
}
