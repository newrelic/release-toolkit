package dependabot_test

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/changelog"
	"github.com/newrelic/release-toolkit/dependabot"
	"github.com/newrelic/release-toolkit/git"
	"github.com/stretchr/testify/assert"
)

type SemverTagsGetterMock struct {
	Hash string
}

func (t *SemverTagsGetterMock) Get() (git.SemverTags, error) {
	version := semver.MustParse("v1.2.3")
	return git.SemverTags{
		Versions: []*semver.Version{version},
		Hashes:   map[*semver.Version]string{version: "a-hash"},
	}, nil
}

func (t *SemverTagsGetterMock) GetLastReleaseHash() (string, error) {
	return "", nil
}

type CommitsGetterMock struct {
	CommitList []git.Commit
}

func (c *CommitsGetterMock) Get(_ string) ([]git.Commit, error) {
	return c.CommitList, nil
}

//nolint:funlen
func TestExtractor_Extract(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name                 string
		commitMessages       []git.Commit
		expectedDependencies []changelog.Dependency
	}{
		{
			name: "Matching_and_not_matching-commits",
			commitMessages: []git.Commit{
				{Message: `Whatever actions/github-script from 1.0.2 to 1.0.4.`},
				{Message: "Non matching"},
				{Message: "chore(deps): bump common-library-2 from v0.0.3 to v0.0.4"},
			},
			expectedDependencies: []changelog.Dependency{
				{Name: "common-library-2", From: semver.MustParse("0.0.3"), To: semver.MustParse("0.0.4")},
			},
		},
		{
			name: "Matching_commits",
			commitMessages: []git.Commit{
				{Message: `build(deps): bump actions/github-script from 4.0.2 to 4.1 whatever
		Bumps [actions/github-script](https://github.com/actions/github-script) from 4.0.2 to 4.1.`},
				{Message: "Bump actions/github-script from 2 to 4.0.2"},
				{Message: "chore(deps): bump github.com/spf13/viper from 1.7.0 to 1.10.1"},
				{Message: "chore(deps): bump github.com/google/go-cmp from 0.5.6 to 0.5.7"},
				{Message: "chore(deps): bump alpine from 3.15.4 to 3.16.0"},
				{Message: "chore(deps): bump github.com/newrelic/infra-integrations-sdk from 3.7.1+incompatible to 3.7.2+incompatible"},
				{Message: "chore(deps): bump common-library-2 from v0.0.3 to v0.0.4"},
				{Message: "chore(deps): bump common-library-3/library from v1.0.3 to v1.2.3"},
				{Message: "chore(deps): bump common-library-4 from v9.10.30 to v10.20.30"},
				{Message: "chore(deps): bump common-library-5 from v1.0.2-prerelease+meta to v1.1.2-prerelease+meta"},
				{Message: "chore(deps): bump common-library-6 from v0.9.2+meta to v1.1.2+meta"},
				{Message: "chore(deps): bump common-library-7 from v0.9.2+meta-valid to v1.1.2+meta-valid"},
				{Message: "chore(deps): bump common-library-8 from v0.9.0-alpha to v1.0.0-alpha"},
				{Message: "chore(deps): bump common-library-9 from v0.9.0-alpha.beta.1 to v1.0.0-alpha.beta.1"},
				{Message: "chore(deps): bump common-library-10 from v0.9.0-alpha.1 to v1.0.0-alpha.1"},
				{Message: "chore(deps): bump common-library-11 from v0.9.0-alpha0.valid to v1.0.0-alpha0.valid"},
				{Message: "chore(deps): bump common-library-12 from v0.9.0-alpha.0 to v1.0.0-alpha.0"},
				{Message: "chore(deps): bump common-library-13 from v0.9.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay to v1.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay"},
				{Message: "chore(deps): bump common-library-14 from v1.6.0-rc.1+build.123 to v2.0.0-rc.1+build.123"},
				{Message: "chore(deps): bump common-library-15 from v1.1.0-SNAPSHOT-123 to v1.2.3-SNAPSHOT-123"},
				{Message: "chore(deps): bump common-library-16 from v1.0.0+build.1848 to v2.0.0+build.1848"},
				{Message: "chore(deps): bump common-library-17 from v1.0.1-alpha.1227 to v2.0.1-alpha.1227"},
				{Message: "chore(deps): bump common-library-18 from v0.1.0-alpha+beta to v1.0.0-alpha+beta"},
				{Message: "chore(deps): bump common-library-19 from v1.1.0----RC-SNAPSHOT.12.9.1 to v1.2.3----RC-SNAPSHOT.12.9.1"},
				{Message: "chore(deps): bump common-library-20 from v0.1.0+0.build.1-rc.10000aaa-kk-0.1 to v1.0.0+0.build.1-rc.10000aaa-kk-0.1"},
				{Message: "chore(deps): bump common-library-21 from v888888.999999.999999 to v999999.999999.999999"},
			},
			expectedDependencies: []changelog.Dependency{
				{Name: "actions/github-script", From: semver.MustParse("4.0.2"), To: semver.MustParse("4.1")},
				{Name: "actions/github-script", From: semver.MustParse("2"), To: semver.MustParse("4.0.2")},
				{Name: "github.com/spf13/viper", From: semver.MustParse("1.7.0"), To: semver.MustParse("1.10.1")},
				{Name: "github.com/google/go-cmp", From: semver.MustParse("0.5.6"), To: semver.MustParse("0.5.7")},
				{Name: "alpine", From: semver.MustParse("3.15.4"), To: semver.MustParse("3.16.0")},
				{Name: "github.com/newrelic/infra-integrations-sdk", From: semver.MustParse("3.7.1+incompatible"), To: semver.MustParse("3.7.2+incompatible")},
				{Name: "common-library-2", From: semver.MustParse("0.0.3"), To: semver.MustParse("0.0.4")},
				{Name: "common-library-3/library", From: semver.MustParse("1.0.3"), To: semver.MustParse("1.2.3")},
				{Name: "common-library-4", From: semver.MustParse("9.10.30"), To: semver.MustParse("10.20.30")},
				{Name: "common-library-5", From: semver.MustParse("1.0.2-prerelease+meta"), To: semver.MustParse("1.1.2-prerelease+meta")},
				{Name: "common-library-6", From: semver.MustParse("0.9.2+meta"), To: semver.MustParse("1.1.2+meta")},
				{Name: "common-library-7", From: semver.MustParse("0.9.2+meta-valid"), To: semver.MustParse("1.1.2+meta-valid")},
				{Name: "common-library-8", From: semver.MustParse("0.9.0-alpha"), To: semver.MustParse("1.0.0-alpha")},
				{Name: "common-library-9", From: semver.MustParse("0.9.0-alpha.beta.1"), To: semver.MustParse("1.0.0-alpha.beta.1")},
				{Name: "common-library-10", From: semver.MustParse("0.9.0-alpha.1"), To: semver.MustParse("1.0.0-alpha.1")},
				{Name: "common-library-11", From: semver.MustParse("0.9.0-alpha0.valid"), To: semver.MustParse("1.0.0-alpha0.valid")},
				{Name: "common-library-12", From: semver.MustParse("0.9.0-alpha.0"), To: semver.MustParse("1.0.0-alpha.0")},
				{Name: "common-library-13", From: semver.MustParse("0.9.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay"), To: semver.MustParse("1.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay")},
				{Name: "common-library-14", From: semver.MustParse("1.6.0-rc.1+build.123"), To: semver.MustParse("2.0.0-rc.1+build.123")},
				{Name: "common-library-15", From: semver.MustParse("1.1.0-SNAPSHOT-123"), To: semver.MustParse("1.2.3-SNAPSHOT-123")},
				{Name: "common-library-16", From: semver.MustParse("1.0.0+build.1848"), To: semver.MustParse("2.0.0+build.1848")},
				{Name: "common-library-17", From: semver.MustParse("1.0.1-alpha.1227"), To: semver.MustParse("2.0.1-alpha.1227")},
				{Name: "common-library-18", From: semver.MustParse("0.1.0-alpha+beta"), To: semver.MustParse("v1.0.0-alpha+beta")},
				{Name: "common-library-19", From: semver.MustParse("1.1.0----RC-SNAPSHOT.12.9.1"), To: semver.MustParse("v1.2.3----RC-SNAPSHOT.12.9.1")},
				{Name: "common-library-20", From: semver.MustParse("0.1.0+0.build.1-rc.10000aaa-kk-0.1"), To: semver.MustParse("1.0.0+0.build.1-rc.10000aaa-kk-0.1")},
				{Name: "common-library-21", From: semver.MustParse("888888.999999.999999"), To: semver.MustParse("999999.999999.999999")},
			},
		},
		{
			name: "Matching_commits_with_meta",
			commitMessages: []git.Commit{
				{Message: `build(deps): bump actions/github-script from 4.0.2 to 4.1 (#193)
		Bumps [actions/github-script](https://github.com/actions/github-script) from 4.0.2 to 4.1.`, Author: "dependabot", Hash: "abcda222"},
				{Message: "Bump actions/github-script from 2 to 4.0.2 (#116)", Author: "dependabot", Hash: "abcda222"},
				{Message: "chore(deps): bump github.com/spf13/viper from 1.7.0 to 1.10.1", Author: "dependabot", Hash: "abcda222"},
			},
			expectedDependencies: []changelog.Dependency{
				{
					Name: "actions/github-script",
					From: semver.MustParse("4.0.2"),
					To:   semver.MustParse("4.1"),
					Meta: changelog.EntryMeta{
						Author: "dependabot",
						PR:     "#193",
						Commit: "abcda222",
					},
				},
				{
					Name: "actions/github-script",
					From: semver.MustParse("2"),
					To:   semver.MustParse("4.0.2"),
					Meta: changelog.EntryMeta{
						Author: "dependabot",
						PR:     "#116",
						Commit: "abcda222",
					},
				},
				{
					Name: "github.com/spf13/viper",
					From: semver.MustParse("1.7.0"),
					To:   semver.MustParse("1.10.1"),
					Meta: changelog.EntryMeta{
						Author: "dependabot",
						PR:     "",
						Commit: "abcda222",
					},
				},
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			extractor := dependabot.NewExtractor(&SemverTagsGetterMock{}, &CommitsGetterMock{CommitList: tc.commitMessages})
			dependencies, err := extractor.Extract()
			if err != nil {
				t.Fatalf("Error extracting renovate dependencies: %v", err)
			}

			assert.Equal(t, len(tc.expectedDependencies), len(dependencies))
			for k, dep := range dependencies {
				assert.Equal(t, tc.expectedDependencies[k].Name, dep.Name)
				assert.Equal(t, tc.expectedDependencies[k].To.String(), dep.To.String())
				assert.Equal(t, tc.expectedDependencies[k].Meta, dep.Meta)
			}
		})
	}
}
