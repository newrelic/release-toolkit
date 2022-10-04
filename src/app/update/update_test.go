package update_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/google/go-cmp/cmp"
	"github.com/newrelic/release-toolkit/src/app"
)

//nolint:funlen,paralleltest // urfave/cli cannot be tested concurrently.
func TestRender(t *testing.T) {
	for _, tc := range []struct {
		name     string
		yaml     string
		expected string
		existing string
		args     string
	}{
		{
			name: "Full_Changelog_Now",
			args: "-version v1.2.4",
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
			existing: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should be preserved
			`) + "\n",
			expected: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## v1.2.4 - {NOW}

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

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should be preserved
			`) + "\n",
		},
		{
			name: "Full_Changelog_With_Date",
			args: "-version v1.2.4 -date 1993-09-21",
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
			existing: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should be preserved
			`) + "\n",
			expected: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## v1.2.4 - 1993-09-21

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

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should be preserved
			`) + "\n",
		},
		{
			name: "Changelog_Old_Version_With_Brackets",
			args: "-version v1.2.4 -date 1993-09-21",
			yaml: strings.TrimSpace(`
notes: |-
    ### Important announcement (note)

    This is a release note
changes:
- type: breaking
  message: Support has been removed
dependencies:
- name: foobar
  from: 0.0.1
  to: 0.1.0
			`),
			existing: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## [1.2.3] - 2022-09-20

### Added
- This is in the past and should be preserved
			`) + "\n",
			expected: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## v1.2.4 - 1993-09-21

### Important announcement (note)

This is a release note

### ⚠️️ Breaking changes ⚠️
- Support has been removed

### ⛓️ Dependencies
- Upgraded foobar from 0.0.1 to 0.1.0

## [1.2.3] - 2022-09-20

### Added
- This is in the past and should be preserved
			`) + "\n",
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
			mdFile, err := os.Create(mdPath)
			if err != nil {
				t.Fatalf("Error creating existing changelog for test: %v", err)
			}
			_, _ = mdFile.WriteString(tc.existing)
			_ = mdFile.Close()

			err = app.Run(strings.Fields(fmt.Sprintf("rt -yaml %s update-changelog -markdown %s %s", yamlPath, mdPath, tc.args)))
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

//nolint:paralleltest // urfave/cli cannot be tested concurrently.
func TestRender_Fails_Without_Version(t *testing.T) {
	tDir := t.TempDir()

	app := app.App()

	yamlPath := path.Join(tDir, "changelog.yaml")
	yamlFile, err := os.Create(yamlPath)
	if err != nil {
		t.Fatalf("Error creating yaml for test: %v", err)
	}
	_, _ = yamlFile.WriteString(`held: false`)
	_ = yamlFile.Close()

	mdPath := path.Join(tDir, "changelog.md")
	mdFile, err := os.Create(mdPath)
	if err != nil {
		t.Fatalf("Error creating existing changelog for test: %v", err)
	}
	_ = mdFile.Close()

	err = app.Run([]string{
		"rt", "-yaml", yamlPath, "update-changelog", "-markdown", mdPath,
	})

	if !errors.Is(err, semver.ErrInvalidSemVer) {
		t.Fatalf("App did not return an error with an invalid semver: %v", err)
	}
}

//nolint:paralleltest // urfave/cli cannot be tested concurrently.
func TestRender_Keeps_Backup_File(t *testing.T) {
	tDir := t.TempDir()

	app := app.App()

	yamlPath := path.Join(tDir, "changelog.yaml")
	yamlFile, err := os.Create(yamlPath)
	if err != nil {
		t.Fatalf("Error creating yaml for test: %v", err)
	}
	_, _ = yamlFile.WriteString(`held: false`)
	_ = yamlFile.Close()

	mdPath := path.Join(tDir, "changelog.md")
	mdFile, err := os.Create(mdPath)
	if err != nil {
		t.Fatalf("Error creating existing changelog for test: %v", err)
	}
	_, _ = mdFile.WriteString(`# Changelog
## Unreleased

Yada yada
`)
	_ = mdFile.Close()

	mdContents, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatal(err)
	}
	changelogSum := sha256.Sum256(mdContents)

	err = app.Run(
		[]string{
			"rt", "-yaml", yamlPath, "update-changelog", "-markdown", mdPath, "-version", "v1.2.4",
		})
	if err != nil {
		t.Fatal(err)
	}

	bakPath := mdPath + ".bak"
	bakContents, err := os.ReadFile(bakPath)
	if err != nil {
		t.Fatalf("Error reading %s: %v", bakPath, err)
	}
	bakSum := sha256.Sum256(bakContents)

	if changelogSum != bakSum {
		t.Fatalf(
			"Backup file is not identical to original file: %s != %s",
			hex.EncodeToString(changelogSum[:]), hex.EncodeToString(bakSum[:]),
		)
	}
}
