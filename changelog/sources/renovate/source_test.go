package renovate_test

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/changelog"
	"github.com/newrelic/release-toolkit/changelog/sources/renovate"
	"github.com/newrelic/release-toolkit/git"
	"github.com/stretchr/testify/assert"
)

type tagsVersionGetterMock struct{}

func (t *tagsVersionGetterMock) Versions() ([]*semver.Version, error) {
	version := semver.MustParse("v1.2.3")
	return []*semver.Version{version}, nil
}

func (t *tagsVersionGetterMock) LastVersionHash() (string, error) {
	return "", nil
}

type commitsGetterMock struct {
	commitList []git.Commit
}

// Commits return the list of commits in reverse order, which is like the real commit getter would return them if
// the first commit in the slice was committed first.
func (c *commitsGetterMock) Commits(_ string) ([]git.Commit, error) {
	var commits []git.Commit
	for i := len(c.commitList) - 1; i >= 0; i-- {
		commits = append(commits, c.commitList[i])
	}

	return commits, nil
}

//nolint:funlen
func TestSource_Source(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name                 string
		defaultAuthor        string
		commitMessages       []git.Commit
		expectedDependencies []changelog.Dependency
	}{
		{
			name:          "Matching_and_not_matching-commits",
			defaultAuthor: "renovate[bot] <29139614+renovate[bot]@users.noreply.github.com>",
			commitMessages: []git.Commit{
				{Message: "chore(deps): Another commit message v1.0.4 (#401)"},
				{Message: "Non matching"},
				{Message: "chore(deps): update helm release common-library-3 to v1.2.3"},
				{Message: "chore(deps): update helm release common-library-3 to v1.2.3", Author: "foobar"},
			},
			expectedDependencies: []changelog.Dependency{
				{Name: "common-library-3", To: semver.MustParse("1.2.3")},
			},
		},
		{
			name:          "Matching_commits",
			defaultAuthor: "renovate[bot] <29139614+renovate[bot]@users.noreply.github.com>",
			commitMessages: []git.Commit{
				{Message: `chore(deps): update helm release common-library-1 to v1.0.4

* chore(deps): update helm release common-library to v1.0.4
* Bum chart's version
* fix typo in the common library`},
				{Message: "chore(deps): update helm release common-library-2 to v0.0.4"},
				{Message: "chore(deps): update helm release common-library-3/library to v1.2.3"},
				{Message: "chore(deps): update helm release common-library-4 to v10.20.30"},
				{Message: "chore(deps): update helm release common-library-5 to v1.1.2-prerelease+meta"},
				{Message: "chore(deps): update helm release common-library-6 to v1.1.2+meta"},
				{Message: "chore(deps): update helm release common-library-7 to v1.1.2+meta-valid"},
				{Message: "chore(deps): update helm release common-library-8 to v1.0.0-alpha"},
				{Message: "chore(deps): update helm release common-library-9 to v1.0.0-alpha.beta.1"},
				{Message: "chore(deps): update helm release common-library-10 to v1.0.0-alpha.1"},
				{Message: "chore(deps): update helm release common-library-11 to v1.0.0-alpha0.valid"},
				{Message: "chore(deps): update helm release common-library-12 to v1.0.0-alpha.0"},
				{Message: "chore(deps): update helm release common-library-13 to v1.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay"},
				{Message: "chore(deps): update helm release common-library-14 to v2.0.0-rc.1+build.123"},
				{Message: "chore(deps): update helm release common-library-15 to v1.2.3-SNAPSHOT-123"},
				{Message: "chore(deps): update helm release common-library-16 to v2.0.0+build.1848"},
				{Message: "chore(deps): update helm release common-library-17 to v2.0.1-alpha.1227"},
				{Message: "chore(deps): update helm release common-library-18 to v1.0.0-alpha+beta"},
				{Message: "chore(deps): update helm release common-library-19 to v1.2.3----RC-SNAPSHOT.12.9.1"},
				{Message: "chore(deps): update helm release common-library-20 to v1.0.0+0.build.1-rc.10000aaa-kk-0.1"},
				{Message: "chore(deps): update helm release common-library-21 to v999999.999999.999999"},
				{Message: "update rust crate i-do-not-exist to v1.2.3"},
				{Message: "update noprefix to v1.2.3"},
				{Message: "update name with spaces to v1.2.3"},
				{Message: "update fancy-module to v1.2.3"},
				{Message: "update module to v1.2.3"},
				// From github.com/newrelic/nri-kubernetes.
				{Message: "chore(deps): update newrelic/infrastructure-bundle docker tag to v2.7.6 (#280)"},
				// From github.com/newrelic/infrastructure-bundle
				{Message: "fix(deps): update module github.com/google/go-github/v39 to v39.2.0 (#123)"},
				{Message: "chore(deps): update newrelic/infrastructure docker tag to v1.20.5 (#125)"},
				{Message: "chore(deps): update integrations (#124)"},
				{Message: "chore(deps): update aquasecurity/trivy-action action to v0.0.22 (#127)"},
				{Message: "chore(deps): update dependency newrelic/nri-jmx to v2.6.0 (#129)"},
				{Message: "chore(deps): update github actions to v2 (major) (#178)"},
				{Message: "chore(deps): update github actions to v2.1 (minor) (#179)"},
				{Message: "chore(deps): update github actions to v2.1.1 (patch) (#180)"},
			},
			expectedDependencies: []changelog.Dependency{
				{Name: "common-library-1", To: semver.MustParse("v1.0.4")},
				{Name: "common-library-2", To: semver.MustParse("0.0.4")},
				{Name: "common-library-3/library", To: semver.MustParse("1.2.3")},
				{Name: "common-library-4", To: semver.MustParse("10.20.30")},
				{Name: "common-library-5", To: semver.MustParse("1.1.2-prerelease+meta")},
				{Name: "common-library-6", To: semver.MustParse("1.1.2+meta")},
				{Name: "common-library-7", To: semver.MustParse("1.1.2+meta-valid")},
				{Name: "common-library-8", To: semver.MustParse("1.0.0-alpha")},
				{Name: "common-library-9", To: semver.MustParse("1.0.0-alpha.beta.1")},
				{Name: "common-library-10", To: semver.MustParse("1.0.0-alpha.1")},
				{Name: "common-library-11", To: semver.MustParse("1.0.0-alpha0.valid")},
				{Name: "common-library-12", To: semver.MustParse("1.0.0-alpha.0")},
				{Name: "common-library-13", To: semver.MustParse("1.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay")},
				{Name: "common-library-14", To: semver.MustParse("2.0.0-rc.1+build.123")},
				{Name: "common-library-15", To: semver.MustParse("1.2.3-SNAPSHOT-123")},
				{Name: "common-library-16", To: semver.MustParse("2.0.0+build.1848")},
				{Name: "common-library-17", To: semver.MustParse("2.0.1-alpha.1227")},
				{Name: "common-library-18", To: semver.MustParse("v1.0.0-alpha+beta")},
				{Name: "common-library-19", To: semver.MustParse("v1.2.3----RC-SNAPSHOT.12.9.1")},
				{Name: "common-library-20", To: semver.MustParse("1.0.0+0.build.1-rc.10000aaa-kk-0.1")},
				{Name: "common-library-21", To: semver.MustParse("999999.999999.999999")},
				// Intentional limitation: "rust crate" is not a known manager so the whole string is the dependency name
				{Name: "rust crate i-do-not-exist", To: semver.MustParse("v1.2.3")},
				{Name: "noprefix", To: semver.MustParse("v1.2.3")},
				{Name: "name with spaces", To: semver.MustParse("v1.2.3")},
				{Name: "fancy-module", To: semver.MustParse("v1.2.3")},
				{Name: "module", To: semver.MustParse("v1.2.3")},
				// From github.com/newrelic/nri-kubernetes
				{Name: "newrelic/infrastructure-bundle", To: semver.MustParse("v2.7.6"), Meta: changelog.EntryMeta{PR: "280"}},
				// From github.com/newrelic/infrastructure-bundle
				{Name: "github.com/google/go-github/v39", To: semver.MustParse("v39.2.0"), Meta: changelog.EntryMeta{PR: "123"}},
				{Name: "newrelic/infrastructure", To: semver.MustParse("v1.20.5"), Meta: changelog.EntryMeta{PR: "125"}},
				{Name: "integrations", Meta: changelog.EntryMeta{PR: "124"}},
				{Name: "aquasecurity/trivy-action", To: semver.MustParse("v0.0.22"), Meta: changelog.EntryMeta{PR: "127"}},
				{Name: "newrelic/nri-jmx", To: semver.MustParse("v2.6.0"), Meta: changelog.EntryMeta{PR: "129"}},
				{Name: "github actions", To: semver.MustParse("v2.0.0"), Meta: changelog.EntryMeta{PR: "178"}},
				{Name: "github actions", To: semver.MustParse("v2.1.0"), Meta: changelog.EntryMeta{PR: "179"}},
				{Name: "github actions", To: semver.MustParse("v2.1.1"), Meta: changelog.EntryMeta{PR: "180"}},
			},
		},
		{
			name:          "Matching_commits_with_meta",
			defaultAuthor: "renovate[bot] <29139614+renovate[bot]@users.noreply.github.com>",
			commitMessages: []git.Commit{
				{Message: "chore(deps): update helm release common-library-1 to v1.0.4 (#401)", Hash: "abcda222"},
				{Message: "chore(deps): update helm release common-library-2 to v0.0.4 (#402)", Hash: "abcda222"},
				{Message: "chore(deps): update helm release common-library-3 to v1.2.3", Hash: "abcda222"},
			},
			expectedDependencies: []changelog.Dependency{
				{
					Name: "common-library-1",
					To:   semver.MustParse("v1.0.4"),
					Meta: changelog.EntryMeta{
						PR:     "401",
						Commit: "abcda222",
					},
				},
				{
					Name: "common-library-2",
					To:   semver.MustParse("v0.0.4"),
					Meta: changelog.EntryMeta{
						PR:     "402",
						Commit: "abcda222",
					},
				},
				{
					Name: "common-library-3",
					To:   semver.MustParse("v1.2.3"),
					Meta: changelog.EntryMeta{
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

			for i := range tc.commitMessages {
				if tc.commitMessages[i].Author == "" {
					tc.commitMessages[i].Author = tc.defaultAuthor
				}
			}

			extractor := renovate.NewSource(&tagsVersionGetterMock{}, &commitsGetterMock{commitList: tc.commitMessages})
			cl, err := extractor.Changelog()
			if err != nil {
				t.Fatalf("Error extracting renovate dependencies: %v", err)
			}

			assert.Equal(t, len(tc.expectedDependencies), len(cl.Dependencies))
			for k, dep := range cl.Dependencies {
				assert.Equal(t, tc.expectedDependencies[k].Name, dep.Name)
				if dep.To != nil {
					assert.Equal(t, tc.expectedDependencies[k].To.String(), dep.To.String())
				} else {
					assert.Nil(t, tc.expectedDependencies[k].To)
				}
				assert.Equal(t, tc.expectedDependencies[k].Meta, dep.Meta)
			}
		})
	}
}
