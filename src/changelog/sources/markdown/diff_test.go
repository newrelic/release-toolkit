package markdown_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/newrelic/release-toolkit/src/changelog/sources/markdown"
)

//nolint:funlen
func TestValidator_ValidateDiff(t *testing.T) {
	t.Parallel()

	const base = `# Changelog
This is based on blah blah blah

## Unreleased

### Breaking
- Support has been removed

## v1.2.3 - 2022-11-11

### Enhancements
- This is in the past and should not be included
`

	for _, tc := range []struct {
		name       string
		oldContent string
		newContent string
		expectErr  bool
	}{
		{
			name:       "No_changes",
			oldContent: base,
			newContent: base,
			expectErr:  false,
		},
		{
			name:       "Line_added_in_Unreleased",
			oldContent: base,
			newContent: strings.Replace(base, "- Support has been removed\n", "- Support has been removed\n- Another breaking change\n", 1),
			expectErr:  false,
		},
		{
			name:       "Line_added_in_old_version",
			oldContent: base,
			newContent: strings.Replace(base, "- This is in the past and should not be included\n", "- This is in the past and should not be included\n- Sneaked in here\n", 1),
			expectErr:  true,
		},
		{
			name:       "Line_added_after_last_section",
			oldContent: base,
			newContent: base + "Sneaked in at the end\n",
			expectErr:  true,
		},
		{
			name: "First_release_cut_from_Unreleased_when_no_prior_release_exists",
			oldContent: `# Changelog
This is based on blah blah blah

## Unreleased

### Breaking
- Support has been removed
`,
			newContent: `# Changelog
This is based on blah blah blah

## Unreleased

## v1.0.0 - 2022-12-01

### Breaking
- Support has been removed
`,
			expectErr: false,
		},
		{
			name:       "Line_added_before_Unreleased",
			oldContent: base,
			newContent: strings.Replace(
				base,
				"This is based on blah blah blah\n",
				"This is based on blah blah blah\nSneaked in here\n",
				1,
			),
			expectErr: false,
		},
		{
			name:       "New_L3_section_added_in_Unreleased",
			oldContent: base,
			newContent: strings.Replace(
				base,
				"## v1.2.3",
				"### Enhancements\n- A new enhancement\n\n## v1.2.3",
				1,
			),
			expectErr: false,
		},
		{
			name:       "Unreleased_content_cut_into_new_release_section",
			oldContent: base,
			newContent: `# Changelog
This is based on blah blah blah

## Unreleased

## v1.3.0 - 2022-12-01

### Breaking
- Support has been removed

## v1.2.3 - 2022-11-11

### Enhancements
- This is in the past and should not be included
`,
			expectErr: false,
		},
		{
			name:       "New_release_section_can_differ_from_old_Unreleased",
			oldContent: base,
			newContent: `# Changelog
This is based on blah blah blah

## Unreleased

## v1.3.0 - 2022-12-01

### Breaking
- Support has been removed
- This was added during release, not validated

## v1.2.3 - 2022-11-11

### Enhancements
- This is in the past and should not be included
`,
			expectErr: false,
		},
		{
			name:       "Release_combined_with_tampered_old_version",
			oldContent: base,
			newContent: `# Changelog
This is based on blah blah blah

## Unreleased

## v1.3.0 - 2022-12-01

### Breaking
- Support has been removed

## v1.2.3 - 2022-11-11

### Enhancements
- This is in the past and should not be included
- Sneaked in here
`,
			expectErr: true,
		},
		{
			name:       "No_Unreleased_header",
			oldContent: "# Changelog\n\n## v1.2.3 - 2022-11-11\n\n### Enhancements\n- Old\n",
			newContent: "# Changelog\n\n## v1.2.3 - 2022-11-11\n\n### Enhancements\n- Old\n- New\n",
			expectErr:  false,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			vr, err := markdown.NewValidator(strings.NewReader(tc.newContent))
			if err != nil {
				t.Fatal(err)
			}

			actual := vr.ValidateDiff(tc.oldContent, tc.newContent)

			if !tc.expectErr {
				if len(actual) != 0 {
					t.Fatalf("expected no errors, got %v", actual)
				}
				return
			}

			if len(actual) != 1 || !errors.Is(actual[0], markdown.ErrChangesInReleasedVersions) {
				t.Fatalf("expected a single ErrChangesInReleasedVersions, got %v", actual)
			}
		})
	}
}
