package bumper_test

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/src/bump"
	"github.com/newrelic/release-toolkit/src/bumper"
	"github.com/newrelic/release-toolkit/src/changelog"
)

type testCase struct {
	name      string
	changelog changelog.Changelog
	current   *semver.Version
	expected  *semver.Version
}

type testCaseWithCap struct {
	name          string
	changelog     changelog.Changelog
	current       *semver.Version
	expected      *semver.Version
	entryCap      bump.Type
	dependencyCap bump.Type
}

//nolint:funlen
func TestBumper_Bump(t *testing.T) {
	t.Parallel()

	testCases := []testCase{
		// Test cases that involve only changes
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

		// Test cases that involve only dependencies
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

		// Mixed test cases
		{
			name:     "Enhancement_and_Patch_On_Deps_Minor",
			current:  semver.MustParse("v1.2.3"),
			expected: semver.MustParse("v1.3.0"),
			changelog: changelog.Changelog{
				Changes: []changelog.Entry{
					{Type: changelog.TypeEnhancement},
					{Type: changelog.TypeSecurity},
				},
				Dependencies: []changelog.Dependency{
					{From: semver.MustParse("v9.9.9"), To: semver.MustParse("v9.9.10")},
				},
			},
		},
		{
			name:     "Bugfix_and_Minor_On_Deps_Minor",
			current:  semver.MustParse("v1.2.3"),
			expected: semver.MustParse("v1.3.0"),
			changelog: changelog.Changelog{
				Changes: []changelog.Entry{
					{Type: changelog.TypeBugfix},
				},
				Dependencies: []changelog.Dependency{
					{From: semver.MustParse("v9.9.9"), To: semver.MustParse("v9.10.9")},
				},
			},
		},
	}

	for _, tc := range testCases {
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

//nolint:funlen
func TestBumper_BumpWithCap(t *testing.T) {
	t.Parallel()

	testCases := []testCaseWithCap{
		// Test cases that involve only changes
		{
			name:     "Minor_Change_limit_to_patch",
			current:  semver.MustParse("v1.2.3"),
			expected: semver.MustParse("v1.2.4"),
			changelog: changelog.Changelog{
				Changes: []changelog.Entry{
					{Type: changelog.TypeEnhancement},
					{Type: changelog.TypeSecurity},
				},
			},
			entryCap: bump.Patch,
		},
		{
			name:     "Major_change_limit_to_patch",
			current:  semver.MustParse("v1.2.3"),
			expected: semver.MustParse("v1.2.4"),
			changelog: changelog.Changelog{
				Changes: []changelog.Entry{
					{Type: changelog.TypeBreaking},
				},
			},
			entryCap: bump.Patch,
		},
		{
			name:     "Major_change_limit_to_minor",
			current:  semver.MustParse("v1.2.3"),
			expected: semver.MustParse("v1.3.0"),
			changelog: changelog.Changelog{
				Changes: []changelog.Entry{
					{Type: changelog.TypeBreaking},
				},
			},
			entryCap: bump.Minor,
		},

		// Test cases that involve only dependencies
		{
			name:     "Minor_dep_limited_to_patch",
			current:  semver.MustParse("v1.2.3"),
			expected: semver.MustParse("v1.2.4"),
			changelog: changelog.Changelog{
				Dependencies: []changelog.Dependency{
					{From: semver.MustParse("v9.9.9"), To: semver.MustParse("v9.10.9")},
				},
			},
			dependencyCap: bump.Patch,
		},
		{
			name:     "Major_dep_limited_to_patch",
			current:  semver.MustParse("v1.2.3"),
			expected: semver.MustParse("v1.2.4"),
			changelog: changelog.Changelog{
				Dependencies: []changelog.Dependency{
					{From: semver.MustParse("v9.9.9"), To: semver.MustParse("v10.0.0")},
				},
			},
			dependencyCap: bump.Patch,
		},
		{
			name:     "Major_dep_limited_to_minor",
			current:  semver.MustParse("v1.2.3"),
			expected: semver.MustParse("v1.3.0"),
			changelog: changelog.Changelog{
				Dependencies: []changelog.Dependency{
					{From: semver.MustParse("v9.9.9"), To: semver.MustParse("v10.0.0")},
				},
			},
			dependencyCap: bump.Minor,
		},

		// Mixed test cases
		{
			name:     "Breaking_change_limited_to_patch_and_major_dep_limited_to_minor",
			current:  semver.MustParse("v1.2.3"),
			expected: semver.MustParse("v1.3.0"),
			changelog: changelog.Changelog{
				Changes: []changelog.Entry{
					{Type: changelog.TypeBreaking},
				},
				Dependencies: []changelog.Dependency{
					{From: semver.MustParse("v9.9.9"), To: semver.MustParse("v10.0.0")},
				},
			},
			entryCap:      bump.Patch,
			dependencyCap: bump.Minor,
		},
		{
			name:     "Breaking_change_limited_to_minor_and_major_dep_limited_to_patch",
			current:  semver.MustParse("v1.2.3"),
			expected: semver.MustParse("v1.3.0"),
			changelog: changelog.Changelog{
				Changes: []changelog.Entry{
					{Type: changelog.TypeBreaking},
				},
				Dependencies: []changelog.Dependency{
					{From: semver.MustParse("v9.9.9"), To: semver.MustParse("v10.0.0")},
				},
			},
			entryCap:      bump.Minor,
			dependencyCap: bump.Patch,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			bumper := bumper.New(tc.changelog)
			bumper.EntryCap = tc.entryCap
			bumper.DependencyCap = tc.dependencyCap

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

func TestBumper_BumpSource_InitialVersion(t *testing.T) {
	t.Parallel()
	b := bumper.New(changelog.Changelog{
		Changes: []changelog.Entry{
			{
				Type:    changelog.TypeEnhancement,
				Message: "An enhancement",
			},
		},
	})
	source := mockSource{}

	bumped, err := b.BumpSource(source)
	if err == nil {
		t.Fatalf("Expected bumper to error when source is empty, got %v", bumped)
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
