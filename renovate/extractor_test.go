package renovate_test

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/changelog"
	"github.com/newrelic/release-toolkit/commit"
	"github.com/newrelic/release-toolkit/renovate"
	"github.com/newrelic/release-toolkit/tag"
	"github.com/stretchr/testify/assert"
)

type TagSourceMock struct {
	Hash string
}

func (t *TagSourceMock) Tags() ([]tag.Tag, error) {
	return []tag.Tag{
		{
			Version: semver.MustParse("v1.2.3"),
			Hash:    "a-hash",
		},
	}, nil
}

type CommitSourceMock struct {
	CommitList []commit.Commit
}

func (c *CommitSourceMock) Commits(_ string) ([]commit.Commit, error) {
	return c.CommitList, nil
}

func TestExtractor_Extract(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name                 string
		commitMessages       []commit.Commit
		expectedDependencies []changelog.Dependency
	}{
		{
			name: "Matching_and_not_matching-commits",
			commitMessages: []commit.Commit{
				{Message: `chore(deps): Another commit message v1.0.4 (#401)`},
				{Message: "Non matching"},
				{Message: "chore(deps): update helm release common-library-3 to v1.2.3"},
			},
			expectedDependencies: []changelog.Dependency{
				{Name: "common-library-3", To: semver.MustParse("1.2.3")},
			},
		},
		{
			name: "Matching_commits",
			commitMessages: []commit.Commit{
				{Message: `chore(deps): update helm release common-library-1 to v1.0.4 (#401)

* chore(deps): update helm release common-library to v1.0.4
* Bum chart's version
* fix typo in the common library`},
				{Message: "chore(deps): update helm release common-library-2 to v0.0.4 (#402)"},
				{Message: "chore(deps): update helm release common-library-3 to v1.2.3"},
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
			},
			expectedDependencies: []changelog.Dependency{
				{Name: "common-library-1", To: semver.MustParse("v1.0.4")},
				{Name: "common-library-2", To: semver.MustParse("0.0.4")},
				{Name: "common-library-3", To: semver.MustParse("1.2.3")},
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
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			extractor := renovate.NewExtractor(&TagSourceMock{}, &CommitSourceMock{CommitList: tc.commitMessages})
			dependencies, err := extractor.Extract()
			if err != nil {
				t.Fatalf("Error extracting renovate dependencies: %v", err)
			}

			for k, dep := range dependencies {
				assert.Equal(t, tc.expectedDependencies[k].Name, dep.Name)
				assert.Equal(t, tc.expectedDependencies[k].To.String(), dep.To.String())
			}
		})
	}
}
