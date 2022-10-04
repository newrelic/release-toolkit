package markdown_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/newrelic/release-toolkit/src/changelog/sources/markdown/headingdoc"

	"github.com/newrelic/release-toolkit/src/changelog/sources/markdown"
)

//nolint:funlen
func TestValidator_Validate(t *testing.T) {
	t.Parallel()

	// THIS TEST CARES ABOUT ORDER!
	// Ensure that expected entries are sorted as per supportedTypes.
	for _, tc := range []struct {
		name       string
		markdown   string
		parsingErr error
		expected   []error
	}{
		{
			name: "Wrong_changelog_header",
			markdown: strings.TrimSpace(`
# Wrong-Changelog
This is based on blah blah blah

## Unreleased
`),
			expected: []error{markdown.ErrNoChangelogHeader},
		},
		{
			name: "Multiple_L1_headers",
			markdown: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

# another-Changelog
This blah
`),
			parsingErr: headingdoc.ErrUnsupportedL1Header,
		},
		{
			name: "Level1_header_not_at_top",
			markdown: strings.TrimSpace(`
## Unreleased

# Changelog
This is based on blah blah blah
`),
			parsingErr: headingdoc.ErrUnsupportedL1Header,
		},
		{
			name: "Wrong_unreleased_header",
			markdown: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unre
`),
			expected: []error{markdown.ErrNoUnreleasedL2Header},
		},
		{
			name: "Wrong_level_unreleased_header",
			markdown: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

### Unreleased
`),
			expected: []error{markdown.ErrNoUnreleasedL2Header},
		},
		{
			name: "Invalid_Held",
			markdown: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Held

## Unreleased
`),
			expected: []error{markdown.ErrEmptyHeldHeader},
		},
		{
			name: "Valid_Held",
			markdown: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Held

A good held entry

## Unreleased
`),
			expected: []error{},
		},
		{
			name: "No_Held",
			markdown: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unreleased
`),
			expected: []error{},
		},
		{
			name: "Level 4 version header instead of level 3",
			markdown: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unreleased

#### Important announcement (note)
This is a release note
`),
			expected: []error{markdown.ErrL2WrongChildren},
		},
		{
			name: "Empty_level3_header",
			markdown: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unreleased

### Important announcement (note)
`),
			expected: []error{markdown.ErrL3HeaderEmptyContent},
		},
		{
			name: "Empty_level3_entry_type_header",
			markdown: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unreleased

### Breaking
`),
			expected: []error{markdown.ErrL3HeaderEmptyContent},
		},
		{
			name: "No_itemized_content_level3_header",
			markdown: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unreleased

### Breaking
Support has been removed
`),
			expected: []error{markdown.ErrL3HeaderNoItemizedList},
		},
		{
			name: "Valid_Changelog",
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

## v1.2.3 - 2022-11-11

### Enhancements
- This is in the past and should not be included
`),
			expected: []error{},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			vr, err := markdown.NewValidator(strings.NewReader(tc.markdown))
			if err != nil {
				if errors.Is(err, tc.parsingErr) {
					return
				}
				t.Fatal(err)
			}

			actual := vr.Validate()
			for k, errExpected := range tc.expected {
				if len(actual) <= k {
					t.Fatalf("Validator expected to be %q but was empty", errExpected)
				}
				if !errors.Is(actual[k], errExpected) {
					t.Fatalf("Validator expected to be %q but was %q", errExpected, actual[k])
				}
			}
		})
	}
}
