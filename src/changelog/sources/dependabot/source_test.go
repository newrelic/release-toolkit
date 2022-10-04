package dependabot_test

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/google/go-cmp/cmp"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/changelog/sources/dependabot"
	"github.com/newrelic/release-toolkit/src/git"
	"github.com/newrelic/release-toolkit/src/hack"
)

type tagsVersionGetterMock struct{}

func (t *tagsVersionGetterMock) Versions() ([]*semver.Version, error) {
	version := semver.MustParse("v1.2.3")
	return []*semver.Version{version}, nil
}

func (t *tagsVersionGetterMock) LastVersionHash() (string, error) {
	return "", nil
}

// commitList is a mocked commit source.
type commitList []git.Commit

// Commits return the list of commits in reverse order, which is like the real commit getter would return them if
// the first commit in the slice was committed first.
func (cl commitList) Commits(_ string) ([]git.Commit, error) {
	return cl, nil
}

//nolint:funlen
func TestSource_Source(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		commit   git.Commit
		expected []changelog.Dependency
	}{
		{
			commit: git.Commit{Message: "Whatever actions/github-script from 1.0.2 to 1.0.4."},
		},
		{
			commit: git.Commit{Message: "Non matching"},
		},
		{
			commit:   git.Commit{Message: "chore(deps): bump common-library-2 from v0.0.3 to v0.0.4"},
			expected: []changelog.Dependency{{Name: "common-library-2", From: semver.MustParse("0.0.3"), To: semver.MustParse("0.0.4")}},
		},
		{
			commit: git.Commit{Message: `build(deps): bump actions/github-script from 4.0.2 to 4.1 whatever,
Bumps [actions/github-script](https://github.com/actions/github-script) from 4.0.2 to 4.1.`},
			expected: []changelog.Dependency{{Name: "actions/github-script", From: semver.MustParse("4.0.2"), To: semver.MustParse("4.1")}},
		},
		{
			commit:   git.Commit{Message: "Bump actions/github-script from 2 to 4.0.2"},
			expected: []changelog.Dependency{{Name: "actions/github-script", From: semver.MustParse("2"), To: semver.MustParse("4.0.2")}},
		},
		{
			commit:   git.Commit{Message: "chore(deps): bump github.com/spf13/viper from 1.7.0 to 1.10.1"},
			expected: []changelog.Dependency{{Name: "github.com/spf13/viper", From: semver.MustParse("1.7.0"), To: semver.MustParse("1.10.1")}},
		},
		{
			commit:   git.Commit{Message: "chore(deps): bump github.com/google/go-cmp from 0.5.6 to 0.5.7"},
			expected: []changelog.Dependency{{Name: "github.com/google/go-cmp", From: semver.MustParse("0.5.6"), To: semver.MustParse("0.5.7")}},
		},
		{
			commit:   git.Commit{Message: "chore(deps): bump alpine from 3.15.4 to 3.16.0"},
			expected: []changelog.Dependency{{Name: "alpine", From: semver.MustParse("3.15.4"), To: semver.MustParse("3.16.0")}},
		},
		{
			commit:   git.Commit{Message: "chore(deps): bump github.com/newrelic/infra-integrations-sdk from 3.7.1+incompatible to 3.7.2+incompatible"},
			expected: []changelog.Dependency{{Name: "github.com/newrelic/infra-integrations-sdk", From: semver.MustParse("3.7.1+incompatible"), To: semver.MustParse("3.7.2+incompatible")}},
		},
		{
			commit:   git.Commit{Message: "chore(deps): bump common-library-3/library from v1.0.3 to v1.2.3"},
			expected: []changelog.Dependency{{Name: "common-library-3/library", From: semver.MustParse("1.0.3"), To: semver.MustParse("1.2.3")}},
		},
		{
			commit:   git.Commit{Message: "chore(deps): bump common-library-4 from v9.10.30 to v10.20.30"},
			expected: []changelog.Dependency{{Name: "common-library-4", From: semver.MustParse("9.10.30"), To: semver.MustParse("10.20.30")}},
		},
		{
			commit:   git.Commit{Message: "chore(deps): bump common-library-5 from v1.0.2-prerelease+meta to v1.1.2-prerelease+meta"},
			expected: []changelog.Dependency{{Name: "common-library-5", From: semver.MustParse("1.0.2-prerelease+meta"), To: semver.MustParse("1.1.2-prerelease+meta")}},
		},
		{
			commit:   git.Commit{Message: "chore(deps): bump common-library-6 from v0.9.2+meta to v1.1.2+meta"},
			expected: []changelog.Dependency{{Name: "common-library-6", From: semver.MustParse("0.9.2+meta"), To: semver.MustParse("1.1.2+meta")}},
		},
		{
			commit:   git.Commit{Message: "chore(deps): bump common-library-7 from v0.9.2+meta-valid to v1.1.2+meta-valid"},
			expected: []changelog.Dependency{{Name: "common-library-7", From: semver.MustParse("0.9.2+meta-valid"), To: semver.MustParse("1.1.2+meta-valid")}},
		},
		{
			commit:   git.Commit{Message: "chore(deps): bump common-library-13 from v0.9.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay to v1.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay"},
			expected: []changelog.Dependency{{Name: "common-library-13", From: semver.MustParse("0.9.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay"), To: semver.MustParse("1.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay")}},
		},
		{
			commit:   git.Commit{Message: "chore(deps): bump common-library-14 from v1.6.0-rc.1+build.123 to v2.0.0-rc.1+build.123"},
			expected: []changelog.Dependency{{Name: "common-library-14", From: semver.MustParse("1.6.0-rc.1+build.123"), To: semver.MustParse("2.0.0-rc.1+build.123")}},
		},
		{
			commit:   git.Commit{Message: "chore(deps): bump common-library-15 from v1.1.0-SNAPSHOT-123 to v1.2.3-SNAPSHOT-123"},
			expected: []changelog.Dependency{{Name: "common-library-15", From: semver.MustParse("1.1.0-SNAPSHOT-123"), To: semver.MustParse("1.2.3-SNAPSHOT-123")}},
		},
		{
			commit:   git.Commit{Message: "chore(deps): bump common-library-16 from v1.0.0+build.1848 to v2.0.0+build.1848"},
			expected: []changelog.Dependency{{Name: "common-library-16", From: semver.MustParse("1.0.0+build.1848"), To: semver.MustParse("2.0.0+build.1848")}},
		},
		{
			commit:   git.Commit{Message: "chore(deps): bump common-library-18 from v0.1.0-alpha+beta to v1.0.0-alpha+beta"},
			expected: []changelog.Dependency{{Name: "common-library-18", From: semver.MustParse("0.1.0-alpha+beta"), To: semver.MustParse("v1.0.0-alpha+beta")}},
		},
		{
			commit:   git.Commit{Message: "chore(deps): bump common-library-19 from v1.1.0----RC-SNAPSHOT.12.9.1 to v1.2.3----RC-SNAPSHOT.12.9.1"},
			expected: []changelog.Dependency{{Name: "common-library-19", From: semver.MustParse("1.1.0----RC-SNAPSHOT.12.9.1"), To: semver.MustParse("v1.2.3----RC-SNAPSHOT.12.9.1")}},
		},
		{
			commit:   git.Commit{Message: "chore(deps): bump common-library-20 from v0.1.0+0.build.1-rc.10000aaa-kk-0.1 to v1.0.0+0.build.1-rc.10000aaa-kk-0.1"},
			expected: []changelog.Dependency{{Name: "common-library-20", From: semver.MustParse("0.1.0+0.build.1-rc.10000aaa-kk-0.1"), To: semver.MustParse("1.0.0+0.build.1-rc.10000aaa-kk-0.1")}},
		},
		{
			commit:   git.Commit{Message: "bump really/weird-library-3 from v1 to v2"},
			expected: []changelog.Dependency{{Name: "really/weird-library-3", From: semver.MustParse("v1"), To: semver.MustParse("v2")}},
		},
		{
			commit:   git.Commit{Message: "bump unknown-from to v2 (#123)"},
			expected: []changelog.Dependency{{Name: "unknown-from", To: semver.MustParse("v2"), Meta: changelog.EntryMeta{PR: "123"}}},
		},
		{
			commit:   git.Commit{Message: "bump unknown-to from v1 (#123)"},
			expected: []changelog.Dependency{{Name: "unknown-to", From: semver.MustParse("v1"), Meta: changelog.EntryMeta{PR: "123"}}},
		},
		// From github.com/newrelic/nri-kafka
		{
			commit:   git.Commit{Message: "Bump github.com/golangci/golangci-lint from 1.40.0 to 1.42.1 (#144)"},
			expected: []changelog.Dependency{{Name: "github.com/golangci/golangci-lint", From: semver.MustParse("1.40.0"), To: semver.MustParse("1.42.1"), Meta: changelog.EntryMeta{PR: "144"}}},
		},
		{
			commit:   git.Commit{Message: "Bump github.com/Shopify/sarama from 1.30.1 to 1.31.0 (#152)"},
			expected: []changelog.Dependency{{Name: "github.com/Shopify/sarama", From: semver.MustParse("1.30.1"), To: semver.MustParse("1.31.0"), Meta: changelog.EntryMeta{PR: "152"}}},
		},
		{
			commit:   git.Commit{Message: "Bump github.com/newrelic/infra-integrations-sdk (#150)"},
			expected: []changelog.Dependency{{Name: "github.com/newrelic/infra-integrations-sdk", Meta: changelog.EntryMeta{PR: "150"}}},
		},
		// From github.com/newrelic/nri-kubernetes
		{
			commit:   git.Commit{Message: "build(deps): bump aquasecurity/trivy-action from 0.0.18 to 0.0.19 (#181)"},
			expected: []changelog.Dependency{{Name: "aquasecurity/trivy-action", From: semver.MustParse("0.0.18"), To: semver.MustParse("0.0.19"), Meta: changelog.EntryMeta{PR: "181"}}},
		},
		{
			commit:   git.Commit{Message: "build(deps): bump actions/github-script (#193)"},
			expected: []changelog.Dependency{{Name: "actions/github-script", Meta: changelog.EntryMeta{PR: "193"}}},
		},
		{
			commit:   git.Commit{Message: "build(deps): bump github.com/newrelic/infra-integrations-sdk from 3.6.8+incompatible to 3.7.0+incompatible (#236)"},
			expected: []changelog.Dependency{{Name: "github.com/newrelic/infra-integrations-sdk", From: semver.MustParse("3.6.8+incompatible"), To: semver.MustParse("3.7.0+incompatible"), Meta: changelog.EntryMeta{PR: "236"}}},
		},
		// From github.com/newrelic/release-toolkit
		{
			commit:   git.Commit{Message: "build(deps): bump github.com/urfave/cli/v2 from 2.14.0 to 2.14.1 in /src (#65)"},
			expected: []changelog.Dependency{{Name: "github.com/urfave/cli/v2", From: semver.MustParse("2.14.0"), To: semver.MustParse("2.14.1"), Meta: changelog.EntryMeta{PR: "65"}}},
		},
		{
			commit: git.Commit{Message: `build(deps): bump actions/github-script from 4.0.2 to 4.1 (#193)
		Bumps [actions/github-script](https://github.com/actions/github-script) from 4.0.2 to 4.1.`, Hash: "abcda222"},
			expected: []changelog.Dependency{{
				Name: "actions/github-script",
				From: semver.MustParse("4.0.2"),
				To:   semver.MustParse("4.1"),
				Meta: changelog.EntryMeta{
					PR:     "193",
					Commit: "abcda222",
				},
			}},
		},
		{
			commit: git.Commit{Message: "Bump actions/github-script from 2 to 4.0.2 (#116)", Hash: "abcda222"},
			expected: []changelog.Dependency{{
				Name: "actions/github-script",
				From: semver.MustParse("2"),
				To:   semver.MustParse("4.0.2"),
				Meta: changelog.EntryMeta{
					PR:     "116",
					Commit: "abcda222",
				},
			}},
		},
		{
			commit: git.Commit{Message: "chore(deps): bump github.com/spf13/viper from 1.7.0 to 1.10.1", Hash: "abcda222"},
			expected: []changelog.Dependency{{
				Name: "github.com/spf13/viper",
				From: semver.MustParse("1.7.0"),
				To:   semver.MustParse("1.10.1"),
				Meta: changelog.EntryMeta{
					PR:     "",
					Commit: "abcda222",
				},
			}},
		},
	} {
		tc := tc
		if tc.name == "" {
			tc.name = tc.commit.Message
		}

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			source := dependabot.NewSource(&tagsVersionGetterMock{}, commitList{tc.commit})
			cl, err := source.Changelog()
			if err != nil {
				t.Fatalf("Error extracting dependabot dependencies: %v", err)
			}

			// Hack: Sources may return an empty array, which cmp registers as not equal to `nil`.
			// Here we force it to nil if empty.
			if len(cl.Dependencies) == 0 {
				cl.Dependencies = nil
			}

			diff := cmp.Diff(tc.expected, cl.Dependencies, cmp.Comparer(hack.SemverEquals))
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
