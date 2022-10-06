package markdown_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	cl "github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/changelog/sources/markdown"
)

//nolint:funlen
func TestMarkdown_Changelog(t *testing.T) {
	t.Parallel()

	// THIS TEST CARES ABOUT ORDER!
	// Ensure that expected entries are sorted as per supportedTypes.
	for _, tc := range []struct {
		name     string
		markdown string
		expected *cl.Changelog
	}{
		{
			name: "Parses_Full_Changelog",
			markdown: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unreleased

### Important announcement (note)
This is a release note

### Breaking
- Support has been removed

### Enhancements
- Added this
- Improved that

### Bugfixes
- Fixed a bug that caused the world to end
- Fixed a bug that rm -rf'd /

### Security
- Fixed a security issue that leaked all data

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should not be included
`),
			expected: &cl.Changelog{
				Changes: []cl.Entry{
					{Type: cl.TypeBreaking, Message: "Support has been removed"},
					{Type: cl.TypeSecurity, Message: "Fixed a security issue that leaked all data"},
					{Type: cl.TypeEnhancement, Message: "Added this"},
					{Type: cl.TypeEnhancement, Message: "Improved that"},
					{Type: cl.TypeBugfix, Message: "Fixed a bug that caused the world to end"},
					{Type: cl.TypeBugfix, Message: "Fixed a bug that rm -rf'd /"},
				},
				Notes: "### Important announcement (note)\nThis is a release note\n",
			},
		},
		{
			name: "Parses_Held_L2_Header",
			markdown: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Held

Holding release as it contains massive breaking changes

## Unreleased

### Breaking
- Support has been removed

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should not be included
`),
			expected: &cl.Changelog{
				Held: true,
				Changes: []cl.Entry{
					{Type: cl.TypeBreaking, Message: "Support has been removed"},
				},
			},
		},
		{
			name: "Parses_Held_L3_Header",
			markdown: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unreleased

### Held

Holding release as it contains massive breaking changes

### Breaking
- Support has been removed

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should not be included
`),
			expected: &cl.Changelog{
				Held: true,
				Changes: []cl.Entry{
					{Type: cl.TypeBreaking, Message: "Support has been removed"},
				},
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			src := markdown.New(strings.NewReader(tc.markdown))
			chl, err := src.Changelog()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expected, chl); diff != "" {
				t.Fatalf("Changelog is not the expected one:\n%v", diff)
			}
		})
	}
}
