package render_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/newrelic/release-toolkit/src/app"
)

//nolint:funlen,paralleltest // urfave/cli cannot be tested concurrently.
func TestRender(t *testing.T) {
	for _, tc := range []struct {
		name     string
		yaml     string
		args     string
		expected string
	}{
		{
			name: "Changelog_With_Defaults",
			yaml: strings.TrimSpace(`
notes: |-
    ### Important announcement (note)

    This is a release note
changes:
- type: breaking
  message: Support has been removed
- type: enhancement
  message: New feature has been added
- type: security
  message: Fixed a security issue that leaked all data
dependencies:
- name: foobar
  from: 0.0.1
  to: 0.1.0
			`),
			expected: strings.TrimSpace(`
### Important announcement (note)

This is a release note

### ⚠️️ Breaking changes ⚠️
- Support has been removed

### 🛡️ Security notices
- Fixed a security issue that leaked all data

### 🚀 Enhancements
- New feature has been added

### ⛓️ Dependencies
- Upgraded foobar from 0.0.1 to 0.1.0
			`),
		},
		{
			name: "Changelog_With_Version",
			args: "-version v1.2.3",
			yaml: strings.TrimSpace(`
notes: |-
    ### Important announcement (note)

    This is a release note
changes:
- type: breaking
  message: Support has been removed
- type: enhancement
  message: New feature has been added
- type: security
  message: Fixed a security issue that leaked all data
dependencies:
- name: foobar
  from: 0.0.1
  to: 0.1.0
			`),
			expected: strings.TrimSpace(`
## v1.2.3 - {NOW}

### Important announcement (note)

This is a release note

### ⚠️️ Breaking changes ⚠️
- Support has been removed

### 🛡️ Security notices
- Fixed a security issue that leaked all data

### 🚀 Enhancements
- New feature has been added

### ⛓️ Dependencies
- Upgraded foobar from 0.0.1 to 0.1.0
			`),
		},
		{
			name: "Changelog_With_Version_And_Date",
			args: "-version v1.2.3 -date 1993-09-21",
			yaml: strings.TrimSpace(`
notes: |-
    ### Important announcement (note)

    This is a release note
changes:
- type: breaking
  message: Support has been removed
			`),
			expected: strings.TrimSpace(`
## v1.2.3 - 1993-09-21

### Important announcement (note)

This is a release note

### ⚠️️ Breaking changes ⚠️
- Support has been removed
			`),
		},
	} {
		tc := tc
		//nolint:paralleltest // urfave/cli cannot be tested concurrently.
		t.Run(tc.name, func(t *testing.T) {
			tDir := t.TempDir()

			app := app.App()

			yamlPath := path.Join(tDir, "changelog.yaml")
			yamlFile, err := os.Create(yamlPath)
			if err != nil {
				t.Fatalf("Error creating yaml for test: %v", err)
			}
			_, _ = yamlFile.WriteString(tc.yaml)
			_ = yamlFile.Close()

			mdPath := path.Join(tDir, "changelog.md")

			err = app.Run(strings.Fields(fmt.Sprintf("rt -yaml %s render-changelog -markdown %s %s", yamlPath, mdPath, tc.args)))
			if err != nil {
				t.Fatalf("Error running app: %v", err)
			}

			actual, err := os.ReadFile(mdPath)
			if err != nil {
				t.Fatalf("Error reading MD file: %v", err)
			}

			// The default behavior of the command is to use `now` as the date for the release.
			// As we do want to test this default, we hack test data dynamically.
			tc.expected = strings.ReplaceAll(tc.expected, "{NOW}", time.Now().Format("2006-01-02"))
			if diff := cmp.Diff(tc.expected, string(actual)); diff != "" {
				t.Fatalf("Changelog md is not as expected\n%s", diff)
			}
		})
	}
}
