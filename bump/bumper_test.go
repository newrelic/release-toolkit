package bump_test

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/bump"
	"github.com/newrelic/release-toolkit/changelog"
)

type testCase struct {
	name      string
	changelog changelog.Changelog
	current   *semver.Version
	expected  *semver.Version
}

func TestBumper_Bump(t *testing.T) {
	t.Parallel()

	changesOnlyCases := []testCase{
		{
			name:     "Patch_On_Bugfixes_Only",
			current:  semver.MustParse("v1.2.3"),
			expected: semver.MustParse("v1.2.4"),
			changelog: changelog.Changelog{
				Changes: []changelog.Entry{
					{Type: changelog.TypeBugfix},
				},
			},
		},
		{
			name:     "Minor_On_Enhancements",
			current:  semver.MustParse("v1.2.3"),
			expected: semver.MustParse("v1.3.0"),
			changelog: changelog.Changelog{
				Changes: []changelog.Entry{
					{Type: changelog.TypeEnhancement},
					{Type: changelog.TypeSecurity},
				},
			},
		},
		{
			name:     "Major_On_Breaking",
			current:  semver.MustParse("v1.2.3"),
			expected: semver.MustParse("v2.0.0"),
			changelog: changelog.Changelog{
				Changes: []changelog.Entry{
					{Type: changelog.TypeBreaking},
				},
			},
		},
	}

	depsOnlyCases := []testCase{
		{
			name:     "Patch_On_Deps_Patch",
			current:  semver.MustParse("v1.2.3"),
			expected: semver.MustParse("v1.2.4"),
			changelog: changelog.Changelog{
				Dependencies: []changelog.Dependency{
					{From: semver.MustParse("v9.9.9"), To: semver.MustParse("v9.9.10")},
				},
			},
		},
		{
			name:     "Minor_On_Deps_Minor",
			current:  semver.MustParse("v1.2.3"),
			expected: semver.MustParse("v1.3.0"),
			changelog: changelog.Changelog{
				Dependencies: []changelog.Dependency{
					{From: semver.MustParse("v9.9.9"), To: semver.MustParse("v9.10.9")},
				},
			},
		},
		{
			name:     "Major_On_Deps_Major",
			current:  semver.MustParse("v1.2.3"),
			expected: semver.MustParse("v2.0.0"),
			changelog: changelog.Changelog{
				Dependencies: []changelog.Dependency{
					{From: semver.MustParse("v9.9.9"), To: semver.MustParse("v10.0.0")},
				},
			},
		},
	}

	for _, tc := range append(changesOnlyCases, depsOnlyCases...) {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			bumper := bump.NewBumper(tc.changelog)
			next := bumper.Bump(tc.current)
			if !tc.expected.Equal(next) {
				t.Fatalf("Expected %v, got %v", tc.expected, next)
			}
		})
	}
}
