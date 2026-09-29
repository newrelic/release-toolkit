package docsreleasenotes_test

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/newrelic/release-toolkit/src/app"
)

//nolint:paralleltest // urfave/cli cannot be tested concurrently.
func TestRun(t *testing.T) {
	changelog := strings.TrimSpace(`
# Changelog

## Unreleased

## v1.99.0 - 2026-07-15

### Enhancements
- Add support for shared filesystems

### Bug fixes
- Fix a crash on startup

## v1.98.0 - 2026-07-01

### Enhancements
- Should not appear
	`) + "\n"

	expected := strings.TrimSpace(`
---
subject: Agent Control
releaseDate: '2026-07-15'
version: 1.99.0
features: ['Add support for shared filesystems']
enhancements: []
bugs: ['Fix a crash on startup']
security: []
---

### New features

- Add support for shared filesystems

### Bug fixes

- Fix a crash on startup

For a detailed description of changes, see the [release notes](https://github.com/newrelic/newrelic-agent-control/releases/tag/1.99.0).
	`) + "\n"

	tDir := t.TempDir()

	mdPath := path.Join(tDir, "CHANGELOG.md")
	if err := os.WriteFile(mdPath, []byte(changelog), 0o600); err != nil {
		t.Fatalf("Error writing changelog for test: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting working directory: %v", err)
	}
	if err := os.Chdir(tDir); err != nil {
		t.Fatalf("Error changing to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()

	a := app.App()
	args := []string{
		"rt", "nr-docs-release-notes",
		"-changelog", mdPath,
		"-version", "1.99.0",
		"-subject", "Agent Control",
		"-repo", "newrelic/newrelic-agent-control",
	}

	if err := a.Run(args); err != nil {
		t.Fatalf("Error running app: %v", err)
	}

	outPath := path.Join(tDir, "newrelic-agent-control-1-99-0.mdx")
	actual, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("Error reading MDX file: %v", err)
	}

	if diff := cmp.Diff(expected, string(actual)); diff != "" {
		t.Fatalf("MDX output is not as expected\n%s", diff)
	}
}
