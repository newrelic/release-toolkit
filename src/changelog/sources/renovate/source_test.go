package renovate_test

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/changelog/sources/renovate"
	"github.com/newrelic/release-toolkit/src/git"
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
				// From https://github.com/newrelic/helm-charts/pull/930
				{Message: "Update Bundle dependencies (#930)\n[![Mend Renovate](https://app.renovatebot.com/images/banner.svg)](https://renovatebot.com)\n\nThis PR contains the following updates:\n\n| Package | Update | Change |\n|---|---|---|\n| [newrelic-infra-operator](https://hub.docker.com/r/newrelic/newrelic-infra-operator) ([source](https://togithub.com/newrelic/newrelic-infra-operator)) | patch | `1.0.7` -> `1.0.8` |\n| [newrelic-infrastructure](https://docs.newrelic.com/docs/kubernetes-pixie/kubernetes-integration/get-started/introduction-kubernetes-integration/) ([source](https://togithub.com/newrelic/nri-kubernetes)) | patch | `3.8.3` -> `3.8.8` |\n| [newrelic-k8s-metrics-adapter](https://hub.docker.com/r/newrelic/newrelic-k8s-metrics-adapter) ([source](https://togithub.com/newrelic/newrelic-k8s-metrics-adapter)) | patch | `0.7.8` -> `0.7.11` |\n| [nri-kube-events](https://docs.newrelic.com/docs/integrations/kubernetes-integration/kubernetes-events/install-kubernetes-events-integration) ([source](https://togithub.com/newrelic/nri-kube-events)) | patch | `2.2.7` -> `2.2.8` |\n| [nri-metadata-injection](https://hub.docker.com/r/newrelic/k8s-metadata-injection) ([source](https://togithub.com/newrelic/k8s-metadata-injection)) | patch | `3.0.7` -> `3.0.8` |\n| [nri-prometheus](https://docs.newrelic.com/docs/infrastructure/prometheus-integrations/install-configure-openmetrics/configure-prometheus-openmetrics-integrations/) ([source](https://togithub.com/newrelic/nri-prometheus)) | patch | `2.1.11` -> `2.1.13` |\n\n---\n\n### Configuration\n\n📅 **Schedule**: Branch creation - At any time (no schedule defined), Automerge - At any time (no schedule defined).\n\n🚦 **Automerge**: Disabled by config. Please merge this manually once you are satisfied.\n\n♻ **Rebasing**: Renovate will not automatically rebase this PR, because other commits have been found.\n\n👻 **Immortal**: This PR will be recreated if closed unmerged. Get [config help](https://togithub.com/renovatebot/renovate/discussions) if that's undesired.\n\n---\n\n - [ ] <!-- rebase-check -->If you want to rebase/retry this PR, click this checkbox. ⚠ **Warning**: custom changes will be lost.\n\n---\n\nThis PR has been generated by [Mend Renovate](https://www.mend.io/free-developer-tools/renovate/). View repository job log [here](https://app.renovatebot.com/dashboard#github/newrelic/helm-charts).\n<!--renovate-debug:eyJjcmVhdGVkSW5WZXIiOiIzMi4xOTQuNSIsInVwZGF0ZWRJblZlciI6IjMyLjE5NC41In0=-->\n"},
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
				// From https://github.com/newrelic/helm-charts/pull/930
				{Name: "newrelic-infra-operator", From: semver.MustParse("1.0.7"), To: semver.MustParse("1.0.8"), Meta: changelog.EntryMeta{PR: "930"}},
				{Name: "newrelic-infrastructure", From: semver.MustParse("3.8.3"), To: semver.MustParse("3.8.8"), Meta: changelog.EntryMeta{PR: "930"}},
				{Name: "newrelic-k8s-metrics-adapter", From: semver.MustParse("0.7.8"), To: semver.MustParse("0.7.11"), Meta: changelog.EntryMeta{PR: "930"}},
				{Name: "nri-kube-events", From: semver.MustParse("2.2.7"), To: semver.MustParse("2.2.8"), Meta: changelog.EntryMeta{PR: "930"}},
				{Name: "nri-metadata-injection", From: semver.MustParse("3.0.7"), To: semver.MustParse("3.0.8"), Meta: changelog.EntryMeta{PR: "930"}},
				{Name: "nri-prometheus", From: semver.MustParse("2.1.11"), To: semver.MustParse("2.1.13"), Meta: changelog.EntryMeta{PR: "930"}},
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
