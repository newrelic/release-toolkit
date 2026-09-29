package docsreleasenotes_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/newrelic/release-toolkit/src/docsreleasenotes"
)

const fullChangelog = `# Changelog

All notable changes are documented in this file.

## Unreleased

## v2.0.0 - 2026-07-18

### 🚀 Enhancements
- A new release enhancement that must NOT appear in the 1.99.0 notes

## v1.99.0 - 2026-07-15

### 🚀 Enhancements
- Add support for shared filesystems in on-host agent types
- On-host self-update now skips sub-agent reconciliation while an update is in progress

### ⚠️️ Breaking changes ⚠️
- Remove support for the deprecated ` + "`legacy_mode`" + ` config option

### 🐞 Bug fixes
- Restore ` + "`local_config.yaml`" + ` from the ` + "`.rpmsave`" + ` backup left by a prior uninstall
- Fix a mis-recording of the Instrumented metric
- Supports escaping 'quotes' and ''quotes''

### 🛡️ Security notices
- Bump base image to patch CVE-2026-0001

### ⛓️ Dependencies
- Updated rust crate chrono to 0.4.45
- Updated alpine/helm to v4.2.1

## v1.98.0 - 2026-07-01

### 🚀 Enhancements
- A previous release enhancement that must NOT appear in the 1.99.0 notes
`

const depsOnlyChangelog = `# Changelog

All notable changes are documented in this file.

## Unreleased

## v1.99.0 - 2026-07-15

### ⛓️ Dependencies
- Updated rust crate chrono to 0.4.45
- Updated alpine/helm to v4.2.1

## v1.98.0 - 2026-07-01

### 🚀 Enhancements
- A previous release enhancement that must NOT appear in the 1.99.0 notes
`

func TestRender(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		changelog string
		expected  string
	}{
		{
			name:      "FullChangelog",
			changelog: fullChangelog,
			expected: strings.TrimSpace(`
---
subject: Agent Control
releaseDate: '2026-07-15'
version: 1.99.0
features: ['Add support for shared filesystems in on-host agent types', 'On-host self-update now skips sub-agent reconciliation while an update is in progress']
enhancements: ['Remove support for the deprecated `+"`legacy_mode`"+` config option']
bugs: ['Restore `+"`local_config.yaml`"+` from the `+"`.rpmsave`"+` backup left by a prior uninstall', 'Fix a mis-recording of the Instrumented metric', 'Supports escaping ''quotes'' and ''''quotes''''']
security: ['Bump base image to patch CVE-2026-0001', 'Updated rust crate chrono to 0.4.45', 'Updated alpine/helm to v4.2.1']
---

### New features

- Add support for shared filesystems in on-host agent types
- On-host self-update now skips sub-agent reconciliation while an update is in progress

### Improvements and enhancements

- Remove support for the deprecated `+"`legacy_mode`"+` config option

### Bug fixes

- Restore `+"`local_config.yaml`"+` from the `+"`.rpmsave`"+` backup left by a prior uninstall
- Fix a mis-recording of the Instrumented metric
- Supports escaping 'quotes' and ''quotes''

### Security updates

- Bump base image to patch CVE-2026-0001
- Updated rust crate chrono to 0.4.45
- Updated alpine/helm to v4.2.1

For a detailed description of changes, see the [release notes](https://github.com/newrelic/newrelic-agent-control/releases/tag/1.99.0).
			`) + "\n",
		},
		{
			name:      "DepsOnlyChangelog",
			changelog: depsOnlyChangelog,
			expected: strings.TrimSpace(`
---
subject: Agent Control
releaseDate: '2026-07-15'
version: 1.99.0
features: []
enhancements: []
bugs: []
security: ['Updated rust crate chrono to 0.4.45', 'Updated alpine/helm to v4.2.1']
---

### Security updates

- Updated rust crate chrono to 0.4.45
- Updated alpine/helm to v4.2.1

For a detailed description of changes, see the [release notes](https://github.com/newrelic/newrelic-agent-control/releases/tag/1.99.0).
			`) + "\n",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := docsreleasenotes.Renderer{
				Subject:  "Agent Control",
				Repo:     "newrelic/newrelic-agent-control",
				Sections: docsreleasenotes.DefaultSections,
			}

			buf := &strings.Builder{}
			if err := r.Render(buf, tc.changelog, "1.99.0"); err != nil {
				t.Fatalf("Render() returned error: %v", err)
			}

			if diff := cmp.Diff(tc.expected, buf.String()); diff != "" {
				t.Fatalf("Rendered MDX does not match expected:\n%s", diff)
			}
		})
	}
}

func TestRenderVersionNotFound(t *testing.T) {
	t.Parallel()

	r := docsreleasenotes.Renderer{
		Subject:  "Agent Control",
		Repo:     "newrelic/newrelic-agent-control",
		Sections: docsreleasenotes.DefaultSections,
	}

	buf := &strings.Builder{}
	err := r.Render(buf, fullChangelog, "9.9.9")
	if err == nil {
		t.Fatalf("Render() expected error for missing version, got nil")
	}
}
