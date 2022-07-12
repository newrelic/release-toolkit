package generate_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/newrelic/release-toolkit/app"
)

// TODO: Add test cases for other sources.
func TestGenerate(t *testing.T) {
	t.Parallel()

	tDir := t.TempDir()

	t.Run("Markdown_Only", func(t *testing.T) {
		app := app.App()

		yamlPath := path.Join(tDir, "changelog.yaml")
		yamlFile, err := os.Create(yamlPath)
		if err != nil {
			t.Fatalf("Error creating yaml for test: %v", err)
		}
		_ = yamlFile.Close()

		mdPath := path.Join(tDir, "CHANGELOG.md")
		mdFile, err := os.Create(mdPath)
		if err != nil {
			t.Fatalf("Error creating test markdown source: %v", err)
		}
		_, _ = mdFile.WriteString(mdChangelog)

		err = app.Run(strings.Fields(fmt.Sprintf("rt --changelog %s generate -md %s -renovate false -dependabot false", yamlPath, mdPath)))
		if err != nil {
			t.Fatalf("Error running app: %v", err)
		}

		yaml, err := ioutil.ReadFile(yamlPath)
		if err != nil {
			t.Fatalf("Error reading file created by command: %v", err)
		}
		if diff := cmp.Diff(string(yaml), changelogYaml); diff != "" {
			t.Fatalf("Output YAML is not as expected:\n%s", diff)
		}
	})
}

var (
	//nolint:gochecknoglobals
	mdChangelog = strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unreleased

### Important announcement (note)
This is a release note

### Breaking
- Support has been removed

### Security
- Fixed a security issue that leaked all data

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should not be included
	`)

	//nolint:gochecknoglobals
	changelogYaml = strings.TrimSpace(`
notes: |-
    ### Important announcement (note)
    This is a release note
changes:
    - type: breaking
      message: Support has been removed
    - type: security
      message: Fixed a security issue that leaked all data
dependencies: []
	`) + "\n"
)
