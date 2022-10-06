package merger_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/google/go-cmp/cmp"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/changelog/sources/markdown/merger"
)

// brokenWristwatch gives a correct time twice a day.
func brokenWristwatch() time.Time {
	t, _ := time.Parse("2006-01-02", "1993-09-21")
	return t
}

//nolint:funlen // This is a test.
func TestMerger_Merges(t *testing.T) {
	t.Parallel()

	// A changelog with nearly every possible section.
	fullChangelog := changelog.Changelog{
		Notes: "### Extra\nWe worked really hard on this one, we hope you like it!\n",
		Changes: []changelog.Entry{
			{Type: changelog.TypeBugfix, Message: "Fixed this"},
			{Type: changelog.TypeSecurity, Message: "Leaked that"},
			{Type: changelog.TypeBreaking, Message: "Broken everything"},
			{Type: changelog.TypeEnhancement, Message: "Added feature"},
		},
		Dependencies: []changelog.Dependency{
			{Name: "thanklessly-maintained"},
		},
	}

	// Simple changelog with just one entry.
	simpleChangelog := changelog.Changelog{
		Changes: []changelog.Entry{
			{Type: changelog.TypeBugfix, Message: "Fixed this"},
		},
	}

	for _, tc := range []struct {
		name     string
		ch       changelog.Changelog
		original string
		expected string
	}{
		{
			name: "Full_Changelog",
			ch:   fullChangelog,
			original: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unreleased

### Bugfixes
- This is a section
- That will be deleted

### Security
- Completely wiped down from earth

### This also

Will be wiped

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should be preserved
			`) + "\n",
			expected: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unreleased

## v1.2.4 - 1993-09-21

### Extra
We worked really hard on this one, we hope you like it!

### ⚠️️ Breaking changes ⚠️
- Broken everything

### 🛡️ Security notices
- Leaked that

### 🚀 Enhancements
- Added feature

### 🐞 Bug fixes
- Fixed this

### ⛓️ Dependencies
- Updated thanklessly-maintained

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should be preserved
			`) + "\n",
		},
		{
			name: "Simple_Changelog_With_Held",
			ch:   simpleChangelog,
			original: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Held

This release should be held from automatic releasing as it includes heavy breaking changes

## Unreleased

### Bugfixes
- This is a section
- That will be deleted

### Security
- Completely wiped down from earth

### This also

Will be wiped

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should be preserved
			`) + "\n",
			expected: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unreleased

## v1.2.4 - 1993-09-21

### 🐞 Bug fixes
- Fixed this

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should be preserved
			`) + "\n",
		},
		{
			name: "Simple_Changelog_First_Entry",
			ch:   simpleChangelog,
			original: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unreleased

### Bugfixes
- This is a section
- That will be deleted

### Security
- Completely wiped down from earth
			`) + "\n",
			expected: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unreleased

## v1.2.4 - 1993-09-21

### 🐞 Bug fixes
- Fixed this
			`) + "\n",
			// TODO: If no previous sections existed in the changelog, there will be two newlines at the end.
		},
		{
			name: "Simple_Changelog_No_Previous_With_Unreleased",
			ch:   simpleChangelog,
			original: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unreleased
			`) + "\n",
			expected: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unreleased

## v1.2.4 - 1993-09-21

### 🐞 Bug fixes
- Fixed this
			`) + "\n",
		},
		{
			name: "Simple_Changelog_No_Unreleased",
			ch:   simpleChangelog,
			original: strings.TrimSpace(`
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

### 🐞 Bug fixes
- Fixed this

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should be preserved
			`) + "\n",
		},
		{
			name: "Simple_Changelog_With_Brackets",
			ch:   simpleChangelog,
			original: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## [v1.2.3] - 20YY-DD-MM

### Enhancements
- This is in the past and should be preserved
			`) + "\n",
			expected: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## v1.2.4 - 1993-09-21

### 🐞 Bug fixes
- Fixed this

## [v1.2.3] - 20YY-DD-MM

### Enhancements
- This is in the past and should be preserved
			`) + "\n",
		},
		{
			// This is an edge case where Merger does not behave extremely well.
			// Merger cannot easily handle whether before the place where the new section will be inserted, so we took
			// the decision to assume there will be a space before.
			// This assumption does not hold if there are neither previous entries, nor an "Unreleased" header.
			name: "Simple_Changelog_No_Previous_Without_Unreleased",
			ch:   simpleChangelog,
			original: strings.TrimSpace(`
# Changelog
This is based on blah blah blah
`) + "\n",
			expected: strings.TrimSpace(`
# Changelog
This is based on blah blah blah
## v1.2.4 - 1993-09-21

### 🐞 Bug fixes
- Fixed this
			`) + "\n",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			buf := &strings.Builder{}
			mrg := merger.New(&tc.ch, semver.MustParse("v1.2.4"))
			mrg.ReleasedOn = brokenWristwatch
			err := mrg.Merge(strings.NewReader(tc.original), buf)
			if err != nil {
				t.Fatalf("Merger returned error: %v", err)
			}

			actual := buf.String()
			if diff := cmp.Diff(tc.expected, actual); diff != "" {
				t.Fatalf("Merger did not return the expected changelog:\n%s", diff)
			}
		})
	}
}
