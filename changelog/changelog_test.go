package changelog_test

import (
	"reflect"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/bump"
	"github.com/newrelic/release-toolkit/changelog"
)

func TestChangelog_Merge(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		changelogs []*changelog.Changelog
		expected   *changelog.Changelog
	}{
		{
			name: "Everything",
			changelogs: []*changelog.Changelog{
				{
					Changes: []changelog.Entry{
						{Message: "Change one", Type: changelog.TypeBugfix},
					},
					Dependencies: []changelog.Dependency{
						{Name: "Dependency 1"},
						{Name: "Dependency 2"},
					},
				},
				{
					Held:  true,
					Notes: "### Example notes section\nThey are very important",
					Changes: []changelog.Entry{
						{Message: "Change two", Meta: changelog.EntryMeta{Author: "roobre"}},
						{Message: "Change three"},
					},
					Dependencies: []changelog.Dependency{
						{Name: "Dependency 3", To: semver.MustParse("v3.3.3")},
					},
				},
				{
					Notes: "### Another section\nEven more important",
					Changes: []changelog.Entry{
						{Message: "Change four"},
					},
				},
			},
			expected: &changelog.Changelog{
				Held:  true,
				Notes: "### Example notes section\nThey are very important\n\n### Another section\nEven more important",
				Changes: []changelog.Entry{
					{Message: "Change one", Type: changelog.TypeBugfix},
					{Message: "Change two", Meta: changelog.EntryMeta{Author: "roobre"}},
					{Message: "Change three"},
					{Message: "Change four"},
				},
				Dependencies: []changelog.Dependency{
					{Name: "Dependency 1"},
					{Name: "Dependency 2"},
					{Name: "Dependency 3", To: semver.MustParse("v3.3.3")},
				},
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			first := tc.changelogs[0]
			for _, ch := range tc.changelogs[1:] {
				first.Merge(ch)
			}

			if !reflect.DeepEqual(first.Changes, tc.expected.Changes) {
				t.Fatalf("Entries are not equal")
			}
			if !reflect.DeepEqual(first.Dependencies, tc.expected.Dependencies) {
				t.Fatalf("Dependencies are not equal")
			}
			if first.Notes != tc.expected.Notes {
				t.Fatalf("Notes are not equal")
			}

			if !reflect.DeepEqual(first, tc.expected) {
				t.Fatalf("Changelogs are not equal")
			}
		})
	}
}

func TestChangelog_Merge_Does_Duplicate_Entries(t *testing.T) {
	// Current implementation does not deduplicate entries. This test attest this as intended behavior.
	t.Parallel()

	ch := changelog.Changelog{
		Changes: []changelog.Entry{
			{Message: "Change one", Type: changelog.TypeBugfix},
		},
		Dependencies: []changelog.Dependency{
			{Name: "Dependency 1"},
			{Name: "Dependency 2"},
		},
	}

	ch.Merge(&ch)

	if ch.Dependencies[0] != ch.Dependencies[2] {
		t.Fatalf("Dependencies were deduplicated: %s != %s", ch.Dependencies[0].Name, ch.Dependencies[2].Name)
	}

	if ch.Changes[0] != ch.Changes[1] {
		t.Fatalf("Changes were deduplicated: %s != %s", ch.Changes[0].Message, ch.Changes[1].Message)
	}
}

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
