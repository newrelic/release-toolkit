package git_test

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/git"
	"github.com/stretchr/testify/assert"
)

func TestRepoSemverTagsGetter_GetLastReleaseHash(t *testing.T) {
	t.Parallel()
	repodir := repoWithCommitsAndTags(t,
		"v1.2.3",
		"v1.3.0",
		"v1.4.0",
		"1.5.0",
		"0.1.1.2",
		"helm-chart-1.3.0",
		"helm-chart-1.3.1",
		"2.0.0-beta",
	)

	for _, tc := range []struct {
		name         string
		opts         []git.TagOptionFunc
		expectedHash string
	}{
		{
			name:         "Default_Settings",
			opts:         nil,
			expectedHash: getVersionCommitHash(t, repodir, "2.0.0-beta"),
		},
		{
			name:         "Matching_Leading_v",
			opts:         []git.TagOptionFunc{git.TagsMatching("^v")},
			expectedHash: getVersionCommitHash(t, repodir, "1.4.0"),
		},
		{
			name: "Matching_And_Replacing_Prefix",
			opts: []git.TagOptionFunc{
				git.TagsMatching("^helm-chart-"),
				git.TagsReplacing("helm-chart-", ""),
			},
			expectedHash: getVersionCommitHash(
				t,
				repodir,
				"1.3.1",
				git.TagsMatching("^helm-chart-"),
				git.TagsReplacing("helm-chart-", ""),
			),
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tagsGetter, err := git.NewRepoSemverTagsGetter(repodir, tc.opts...)
			if err != nil {
				t.Fatalf("Error creating git source: %v", err)
			}

			hash, err := tagsGetter.GetLastReleaseHash()
			if err != nil {
				t.Fatalf("Error creating git source: %v", err)
			}

			assert.Equal(t, tc.expectedHash, hash, "Reported hasg does not match")
		})
	}
}

func getVersionCommitHash(t *testing.T, repodir, version string, opts ...git.TagOptionFunc) string {
	tagsGetter, err := git.NewRepoSemverTagsGetter(repodir, opts...)
	if err != nil {
		t.Fatalf("Error creating git source: %v", err)
	}

	tags, err := tagsGetter.Get()
	if err != nil {
		t.Fatalf("Error fetching tags: %v", err)
	}

	var tagVersion *semver.Version
	for _, v := range tags.Versions {
		if v.String() == version {
			tagVersion = v
		}
	}

	return tags.Hashes[tagVersion]
}
