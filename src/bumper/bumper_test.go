package bumper_test

import (
	"errors"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/src/bumper"
	"github.com/newrelic/release-toolkit/src/changelog"
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

			bumper := bumper.New(tc.changelog)
			next, err := bumper.Bump(tc.current)
			if err != nil {
				t.Fatalf("Bumping version: %v", err)
			}

			if !tc.expected.Equal(next) {
				t.Fatalf("Expected %v, got %v", tc.expected, next)
			}
		})
	}
}

func TestBumper_BumpSource_Bumps(t *testing.T) {
	t.Parallel()

	c := changelog.Changelog{
		Changes: []changelog.Entry{
			{Type: changelog.TypeEnhancement},
			{Type: changelog.TypeSecurity},
		},
	}

	b := bumper.New(c)
	source := mockSource{
		"v1.2.3",
		"v3.4.5",
		"v2.3.4",
	}

	bumped, err := b.BumpSource(source)
	if err != nil {
		t.Fatalf("Bumping: %v", err)
	}

	expected := semver.MustParse("v3.5.0")
	if !expected.Equal(bumped) {
		t.Fatalf("Expected %v, got %v", expected, bumped)
	}
}

func TestBumper_BumpSource_Errors(t *testing.T) {
	t.Parallel()

	c := changelog.Changelog{
		Changes: []changelog.Entry{
			{Type: changelog.TypeEnhancement},
			{Type: changelog.TypeSecurity},
		},
	}

	b := bumper.New(c)
	source := mockSource{}

	if _, err := b.BumpSource(source); !errors.Is(err, bumper.ErrNoTags) {
		t.Fatalf("Expected bump.ErrNoTags, got %v", err)
	}
}

type mockSource []string

func (m mockSource) Versions() ([]*semver.Version, error) {
	versions := make([]*semver.Version, 0, len(m))
	for _, v := range m {
		versions = append(versions, semver.MustParse(v))
	}

	return versions, nil
}
