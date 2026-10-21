package generate_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/newrelic/release-toolkit/src/app"
	"github.com/newrelic/release-toolkit/src/changelog"
	"gopkg.in/yaml.v3"
)

const markdownWithOneEntry = `# Changelog
This is based on blah blah blah

## Unreleased

### enhancement
- Added from markdown

## v1.2.3 - 2022-11-11

### Enhancements
- This is in the past and should not be included
`

func writeFile(t *testing.T, filePath, content string) {
	t.Helper()

	if err := os.MkdirAll(path.Dir(filePath), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func runGenerate(t *testing.T, dir, args string) (*changelog.Changelog, error) {
	t.Helper()

	yamlPath := path.Join(dir, "changelog.yaml")
	mdPath := path.Join(dir, "CHANGELOG.md")

	err := app.App().Run(strings.Fields(fmt.Sprintf(
		"rt --yaml %s generate-yaml -git-root %s -renovate=false -dependabot=false -markdown %s %s",
		yamlPath, dir, mdPath, args,
	)))
	if err != nil {
		return nil, fmt.Errorf("running app: %w", err)
	}

	content, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("Error reading file created by command: %v", err)
	}

	ch := &changelog.Changelog{}
	if err := yaml.Unmarshal(content, ch); err != nil {
		t.Fatalf("Error parsing the generated yaml: %v", err)
	}

	return ch, nil
}

// The change files are gathered in addition to the markdown file, which keeps working as usual.
//
//nolint:paralleltest // urfave/cli cannot be tested concurrently.
func TestGenerate_Changes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, path.Join(dir, "CHANGELOG.md"), markdownWithOneEntry)
	writeFile(t, path.Join(dir, ".changelog/fix-a-bug.yaml"), "type: bugfix\nmessage: Fixed by a change file\n")
	writeFile(t, path.Join(dir, ".changelog/new-feature.yaml"), "type: Enhancements\nmessage: Added by a change file\n")
	writeFile(t, path.Join(dir, ".changelog/TEMPLATE.yaml"), "type:\nmessage:\n")

	actual, err := runGenerate(t, dir, "-changes-dir "+path.Join(dir, ".changelog"))
	if err != nil {
		t.Fatalf("Error running app: %v", err)
	}

	expected := &changelog.Changelog{
		Notes: "",
		Changes: []changelog.Entry{
			{Type: changelog.TypeEnhancement, Message: "Added from markdown"},
			{Type: changelog.TypeBugfix, Message: "Fixed by a change file"},
			{Type: changelog.TypeEnhancement, Message: "Added by a change file"},
		},
		Dependencies: []changelog.Dependency{},
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("Generated changelog is not as expected: %s", diff)
	}
}

func messages(ch *changelog.Changelog) []string {
	out := make([]string, 0, len(ch.Changes))
	for _, entry := range ch.Changes {
		out = append(out, entry.Message)
	}

	return out
}

//nolint:paralleltest // urfave/cli cannot be tested concurrently.
func TestGenerate_ChangesDirectory(t *testing.T) {
	for _, tc := range []struct {
		name     string
		args     string
		expected []string
	}{
		{
			name:     "Default_Directory",
			expected: []string{"Added from markdown", "From the default directory"},
		},
		{
			name:     "Specified_Directory_Replaces_Default",
			args:     "-changes-dir other",
			expected: []string{"Added from markdown", "From another directory"},
		},
	} {
		//nolint:paralleltest
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, path.Join(dir, "CHANGELOG.md"), markdownWithOneEntry)
			writeFile(t, path.Join(dir, ".changelog/fix-a-bug.yaml"), "type: bugfix\nmessage: From the default directory\n")
			writeFile(t, path.Join(dir, "other/another.yaml"), "type: bugfix\nmessage: From another directory\n")
			chdir(t, dir)

			actual, err := runGenerate(t, dir, tc.args)
			if err != nil {
				t.Fatalf("Error running app: %v", err)
			}

			if diff := cmp.Diff(tc.expected, messages(actual)); diff != "" {
				t.Fatalf("Generated entries are not as expected: %s", diff)
			}
		})
	}
}

//nolint:paralleltest // urfave/cli cannot be tested concurrently.
func TestGenerate_ChangesCanBeDisabled(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, path.Join(dir, "CHANGELOG.md"), markdownWithOneEntry)
	writeFile(t, path.Join(dir, ".changelog/fix-a-bug.yaml"), "type: bugfix\nmessage: From the default directory\n")
	chdir(t, dir)

	actual, err := runGenerate(t, dir, "-changes-dir=")
	if err != nil {
		t.Fatalf("Error running app: %v", err)
	}

	if diff := cmp.Diff([]string{"Added from markdown"}, messages(actual)); diff != "" {
		t.Fatalf("Generated entries are not as expected: %s", diff)
	}
}

//nolint:paralleltest // urfave/cli cannot be tested concurrently.
func TestGenerate_ChangesDirectoryMayNotExist(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, path.Join(dir, "CHANGELOG.md"), markdownWithOneEntry)

	actual, err := runGenerate(t, dir, "-changes-dir "+path.Join(dir, ".changelog"))
	if err != nil {
		t.Fatalf("Error running app: %v", err)
	}

	if len(actual.Changes) != 1 {
		t.Fatalf("Expected only the entry of the markdown but got: %+v", actual.Changes)
	}
}

//nolint:paralleltest // urfave/cli cannot be tested concurrently.
func TestGenerate_InvalidChangeFails(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, path.Join(dir, "CHANGELOG.md"), markdownWithOneEntry)
	writeFile(t, path.Join(dir, ".changelog/typo.yaml"), "type: bugfix\nmsg: Fixed by a change file\n")

	_, err := runGenerate(t, dir, "-changes-dir "+path.Join(dir, ".changelog"))
	if err == nil || !strings.Contains(err.Error(), "typo.yaml") {
		t.Fatalf("Expected an error naming the invalid change file but got: %v", err)
	}
}

// chdir changes the working directory for the rest of the test.
func chdir(t *testing.T, dir string) {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting working directory: %v", err)
	}

	if err = os.Chdir(dir); err != nil {
		t.Fatalf("Error changing working directory: %v", err)
	}

	t.Cleanup(func() { _ = os.Chdir(previous) })
}
