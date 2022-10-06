package renderer_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/google/go-cmp/cmp"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/changelog/renderer"
)

// brokenWristwatch gives a correct time twice a day.
func brokenWristwatch() time.Time {
	t, _ := time.Parse("2006-01-02", "1993-09-21")
	return t
}

//nolint:funlen
func TestRenderer_Render(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		changelog changelog.Changelog
		date      func() time.Time
		version   *semver.Version
		expected  string
	}{
		{
			name:    "Full_Changelog",
			date:    brokenWristwatch,
			version: semver.MustParse("0.0.0"),
			changelog: changelog.Changelog{
				Notes: `
### Relevant notes

I am a note!
`,
				Changes: []changelog.Entry{
					{
						Type:    changelog.TypeBreaking,
						Message: "Extremely scary breaking change",
						Meta: changelog.EntryMeta{
							Author: "@roobre",
							PR:     "#1",
						},
					},
					{
						Type:    changelog.TypeBugfix,
						Message: "Something was fixed",
						Meta: changelog.EntryMeta{
							Author: "@roobre",
							Commit: "abad1dea",
						},
					},
					{
						Type:    changelog.TypeEnhancement,
						Message: "Exciting new feature!",
						Meta: changelog.EntryMeta{
							Author: "@roobre",
							PR:     "#69",
							Commit: "abad1dea",
						},
					},
					{
						Type:    changelog.TypeSecurity,
						Message: "OOPSIE WOOPSIE!! Uwu We made a fucky wucky!! A wittle fucko boingo! The code monkeys at our headquarters are working VEWY HAWD to fix this!",
					},
				},
				Dependencies: []changelog.Dependency{
					{
						Name: "kubernetes",
						From: semver.MustParse("v99.9.9"),
						To:   semver.MustParse("v100.0.0"),
					},
					{
						Name: "etcd",
						From: semver.MustParse("v99.9.9"),
					},
					{
						Name: "logrus",
						To:   semver.MustParse("v100.0.0"),
					},
				},
			},
			expected: strings.TrimSpace(`
## v0.0.0 - 1993-09-21

### Relevant notes

I am a note!

### ⚠️️ Breaking changes ⚠️
- Extremely scary breaking change, by @roobre (#1)

### 🛡️ Security notices
- OOPSIE WOOPSIE!! Uwu We made a fucky wucky!! A wittle fucko boingo! The code monkeys at our headquarters are working VEWY HAWD to fix this!

### 🚀 Enhancements
- Exciting new feature!, by @roobre (#69)

### 🐞 Bug fixes
- Something was fixed, by @roobre (abad1dea)

### ⛓️ Dependencies
- Upgraded kubernetes from v99.9.9 to v100.0.0
- Updated etcd from v99.9.9
- Updated logrus to v100.0.0
`),
		},
		{
			name:    "Full_Changelog_Without_Notes",
			date:    brokenWristwatch,
			version: semver.MustParse("0.0.0"),
			changelog: changelog.Changelog{
				Changes: []changelog.Entry{
					{
						Type:    changelog.TypeBreaking,
						Message: "Extremely scary breaking change",
						Meta: changelog.EntryMeta{
							Author: "@roobre",
							PR:     "#1",
						},
					},
					{
						Type:    changelog.TypeBugfix,
						Message: "Something was fixed",
						Meta: changelog.EntryMeta{
							Author: "@roobre",
							Commit: "abad1dea",
						},
					},
					{
						Type:    changelog.TypeEnhancement,
						Message: "Exciting new feature!",
						Meta: changelog.EntryMeta{
							Author: "@roobre",
							PR:     "#69",
							Commit: "abad1dea",
						},
					},
					{
						Type:    changelog.TypeSecurity,
						Message: "OOPSIE WOOPSIE!! Uwu We made a fucky wucky!! A wittle fucko boingo! The code monkeys at our headquarters are working VEWY HAWD to fix this!",
					},
				},
				Dependencies: []changelog.Dependency{
					{
						Name: "kubernetes",
						From: semver.MustParse("v99.9.9"),
						To:   semver.MustParse("v100.0.0"),
					},
					{
						Name: "etcd",
						From: semver.MustParse("v99.9.9"),
					},
					{
						Name: "logrus",
						To:   semver.MustParse("v100.0.0"),
					},
				},
			},
			expected: strings.TrimSpace(`
## v0.0.0 - 1993-09-21

### ⚠️️ Breaking changes ⚠️
- Extremely scary breaking change, by @roobre (#1)

### 🛡️ Security notices
- OOPSIE WOOPSIE!! Uwu We made a fucky wucky!! A wittle fucko boingo! The code monkeys at our headquarters are working VEWY HAWD to fix this!

### 🚀 Enhancements
- Exciting new feature!, by @roobre (#69)

### 🐞 Bug fixes
- Something was fixed, by @roobre (abad1dea)

### ⛓️ Dependencies
- Upgraded kubernetes from v99.9.9 to v100.0.0
- Updated etcd from v99.9.9
- Updated logrus to v100.0.0
`),
		},
		{
			name:    "Changelog_Without_Some_Entry_Types",
			date:    brokenWristwatch,
			version: semver.MustParse("0.0.0"),
			changelog: changelog.Changelog{
				Changes: []changelog.Entry{
					{
						Type:    changelog.TypeBugfix,
						Message: "Fixed a bug that was causing everything to explode",
						Meta: changelog.EntryMeta{
							Author: "@roobre",
							PR:     "#1337",
						},
					},
				},
				Dependencies: []changelog.Dependency{
					{
						Name: "a totally legit dependency, not malicious at all",
					},
				},
			},
			expected: strings.TrimSpace(`
## v0.0.0 - 1993-09-21

### 🐞 Bug fixes
- Fixed a bug that was causing everything to explode, by @roobre (#1337)

### ⛓️ Dependencies
- Updated a totally legit dependency, not malicious at all
`),
		},
		{
			name:    "Changelog_Only_Bugfix",
			date:    brokenWristwatch,
			version: semver.MustParse("0.0.0"),
			changelog: changelog.Changelog{
				Changes: []changelog.Entry{
					{
						Type:    changelog.TypeBugfix,
						Message: "Fixed a bug that was causing everything to explode",
						Meta: changelog.EntryMeta{
							Author: "@roobre",
							PR:     "#1337",
						},
					},
				},
			},
			expected: strings.TrimSpace(`
## v0.0.0 - 1993-09-21

### 🐞 Bug fixes
- Fixed a bug that was causing everything to explode, by @roobre (#1337)
`),
		},
		{
			name: "Without_Version",
			date: brokenWristwatch,
			changelog: changelog.Changelog{
				Changes: []changelog.Entry{
					{
						Type:    changelog.TypeBugfix,
						Message: "Fixed a bug that was causing everything to explode",
						Meta: changelog.EntryMeta{
							Author: "@roobre",
							PR:     "#1337",
						},
					},
				},
				Dependencies: []changelog.Dependency{
					{
						Name: "a totally legit dependency, not malicious at all",
					},
				},
			},
			expected: strings.TrimSpace(`
### 🐞 Bug fixes
- Fixed a bug that was causing everything to explode, by @roobre (#1337)

### ⛓️ Dependencies
- Updated a totally legit dependency, not malicious at all
`),
		},
		{
			name:    "Without_Date",
			version: semver.MustParse("0.0.0"),
			changelog: changelog.Changelog{
				Changes: []changelog.Entry{
					{
						Type:    changelog.TypeBugfix,
						Message: "Fixed a bug that was causing everything to explode",
						Meta: changelog.EntryMeta{
							Author: "@roobre",
							PR:     "#1337",
						},
					},
				},
				Dependencies: []changelog.Dependency{
					{
						Name: "a totally legit dependency, not malicious at all",
					},
				},
			},
			expected: strings.TrimSpace(`
## v0.0.0

### 🐞 Bug fixes
- Fixed a bug that was causing everything to explode, by @roobre (#1337)

### ⛓️ Dependencies
- Updated a totally legit dependency, not malicious at all
`),
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := renderer.New(&tc.changelog)
			r.ReleasedOn = tc.date
			r.Next = tc.version

			buf := &strings.Builder{}
			err := r.Render(buf)
			if err != nil {
				t.Fatalf("Rendering changelog: %v", err)
			}
			if diff := cmp.Diff(tc.expected, buf.String()); diff != "" {
				t.Fatalf("Output format is not as expected:\n%s", diff)
			}
		})
	}
}
