package changelog_test

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/bump"
	"github.com/newrelic/release-toolkit/changelog"
)

func TestDependency_BumpType_Handles_Nils(t *testing.T) {
	t.Parallel()

	d := changelog.Dependency{
		From: nil,
		To:   semver.MustParse("v1.2.3"),
	}

	if d.BumpType() != bump.Patch {
		t.Fatalf("Expected patch bump from nil")
	}

	d = changelog.Dependency{
		From: semver.MustParse("v1.2.3"),
		To:   nil,
	}

	if d.BumpType() != bump.Patch {
		t.Fatalf("Expected patch bump to nil")
	}
}

func TestDependency_Change(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		expected string
		d        changelog.Dependency
	}{
		{
			expected: "Upgraded",
			d: changelog.Dependency{
				From: semver.MustParse("v1.2.3"),
				To:   semver.MustParse("v1.2.4"),
			},
		},
		{
			expected: "Downgraded",
			d: changelog.Dependency{
				From: semver.MustParse("v1.2.3"),
				To:   semver.MustParse("v1.1.9"),
			},
		},
		{
			expected: "Updated",
			d: changelog.Dependency{
				From: semver.MustParse("v1.2.3"),
				To:   semver.MustParse("v1.2.3"),
			},
		},
		{
			expected: "Updated",
			d: changelog.Dependency{
				From: nil,
				To:   semver.MustParse("v1.2.3"),
			},
		},
		{
			expected: "Updated",
			d: changelog.Dependency{
				From: semver.MustParse("v1.2.3"),
				To:   nil,
			},
		},
		{
			expected: "Updated",
			d: changelog.Dependency{
				From: nil,
				To:   nil,
			},
		},
	} {
		if actual := tc.d.Change(); actual != tc.expected {
			t.Fatalf("Expected %q for %v -> %v, got %v", tc.expected, tc.d.From, tc.d.To, actual)
		}
	}
}
