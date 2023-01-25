package git_test

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/newrelic/release-toolkit/src/git"
	"github.com/stretchr/testify/assert"
)

func repoWithTags(t *testing.T, tags ...string) string {
	t.Helper()

	dir := t.TempDir()

	cmds := []string{
		// Set default branch name to `master`, as tests assume that later.
		"git init --initial-branch master",
		"git config user.email test@user.tld",
		"git config user.name Test",
		"git config commit.gpgsign false",
		"touch a",
		"git add a",
		"git commit -m test",
	}

	for _, t := range tags {
		cmds = append(cmds, fmt.Sprintf("git tag %s", t))
	}

	executeCMDs(t, cmds, dir)

	return dir
}

func executeCMDs(t *testing.T, cmds []string, dir string) {
	t.Helper()

	for _, cmdline := range cmds {
		cmdparts := strings.Fields(cmdline)
		//nolint:gosec // This is a test, we trust hardcoded input.
		cmd := exec.Command(cmdparts[0], cmdparts[1:]...)
		cmd.Dir = dir

		out := strings.Builder{}
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			t.Errorf("%s output:\n%s", cmdline, out.String())
			t.Fatalf("Error bootstraping test git repo: %v", err)
		}
	}
}

func TestTagSource_Versions(t *testing.T) {
	t.Parallel()

	repodir := repoWithTags(t,
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
		name          string
		tagOpts       []git.TagOptionFunc
		tagSourceOpts []git.TagSourceOptionFunc
		expectedTags  []string
	}{
		{
			name: "Default_Settings",
			expectedTags: []string{
				"2.0.0-beta",
				"1.5.0",
				"1.4.0",
				"1.3.0",
				"1.2.3",
			},
		},
		{
			name:    "Matching_Leading_v",
			tagOpts: []git.TagOptionFunc{git.TagsMatchingRegex("^v")},
			expectedTags: []string{
				"1.4.0",
				"1.3.0",
				"1.2.3",
			},
		},
		{
			name: "Matching_And_Replacing_Prefix",
			tagOpts: []git.TagOptionFunc{
				git.TagsMatchingRegex("^helm-chart-"),
			},
			tagSourceOpts: []git.TagSourceOptionFunc{
				git.TagSourceReplacing("helm-chart-", ""),
			},
			expectedTags: []string{
				"1.3.1",
				"1.3.0",
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tagsGetter, err := git.NewRepoTagsGetter(repodir, tc.tagOpts...)
			if err != nil {
				t.Fatalf("Error creating git source: %v", err)
			}

			src := git.NewTagsSource(tagsGetter, tc.tagSourceOpts...)

			versions, err := src.Versions()
			if err != nil {
				t.Fatalf("Error fetching tags: %v", err)
			}

			strVersions := make([]string, 0, len(versions))
			for _, v := range versions {
				strVersions = append(strVersions, v.String())
			}

			assert.ElementsMatchf(t, tc.expectedTags, strVersions, "Reported tags do not match")
		})
	}
}

func TestTagSource_DifferentBranch(t *testing.T) {
	t.Parallel()

	repodir := repoWithTags(t,
		"v1.2.3",
		"v1.3.0",
		"v1.4.0",
		"1.5.0",
		"0.1.1.2",
		"helm-chart-1.3.0",
		"helm-chart-1.3.1",
		"2.0.0-beta",
	)

	expectedTags := []string{"1.4.0", "1.3.0", "1.2.3"}

	executeCMDs(t, []string{
		// we start again from the root, we create a different branch and we check if
		// the tags from the other branch are still considered
		"git checkout -b different/branch",
		"touch b",
		"git add b",
		"git commit -m test",
		"git tag v9.9.9",
		"git checkout master",
	}, repodir)

	commitsGetter := git.NewRepoCommitsGetter(repodir)
	tagOps := []git.TagOptionFunc{git.TagsMatchingRegex("^v"), git.TagsMatchingCommits(commitsGetter)}

	tagsGetter, err := git.NewRepoTagsGetter(repodir, tagOps...)
	if err != nil {
		t.Fatalf("Error creating git source: %v", err)
	}

	src := git.NewTagsSource(tagsGetter)

	versions, err := src.Versions()
	if err != nil {
		t.Fatalf("Error fetching tags: %v", err)
	}

	strVersions := make([]string, 0, len(versions))
	for _, v := range versions {
		strVersions = append(strVersions, v.String())
	}

	assert.ElementsMatchf(t, expectedTags, strVersions, "Reported tags do not match")
}

func TestRepoTagsSource_LastVersionHash(t *testing.T) {
	t.Parallel()
	repodir := repoWithCommitsAndTags(t,
		testCommitTag{"v1.2.3", []string{"v1.2.3"}},
		testCommitTag{"v1.3.0", []string{"v1.3.0"}},
		testCommitTag{"v1.4.0", []string{"v1.4.0"}},
		testCommitTag{"1.5.0", []string{"1.5.0"}},
		testCommitTag{"0.1.1.2", []string{"0.1.1.2"}},
		testCommitTag{"helm-chart-1.3.0", []string{"helm-chart-1.3.0"}},
		testCommitTag{"helm-chart-1.3.1", []string{"helm-chart-1.3.1"}},
		testCommitTag{"2.0.0-beta", []string{"2.0.0-beta"}},
	)

	for _, tc := range []struct {
		name          string
		tagOpts       []git.TagOptionFunc
		tagSourceOpts []git.TagSourceOptionFunc
		expectedHash  string
		expectedErr   error
	}{
		{
			name:         "Default_Settings",
			expectedHash: getVersionCommitHash(t, repodir, "2.0.0-beta"),
		},
		{
			name: "Matching_Leading_v",
			tagOpts: []git.TagOptionFunc{
				git.TagsMatchingRegex("^v"),
			},
			expectedHash: getVersionCommitHash(t, repodir, "v1.4.0"),
		},
		{
			name: "Matching_And_Replacing_Prefix",
			tagOpts: []git.TagOptionFunc{
				git.TagsMatchingRegex("^helm-chart-"),
			},
			tagSourceOpts: []git.TagSourceOptionFunc{
				git.TagSourceReplacing("helm-chart-", ""),
			},
			expectedHash: getVersionCommitHash(
				t,
				repodir,
				"helm-chart-1.3.1",
				git.TagsMatchingRegex("^helm-chart-"),
			),
		},
		{
			name: "No_Versions_Found",
			tagOpts: []git.TagOptionFunc{
				git.TagsMatchingRegex("^nonexistent-"),
			},
			tagSourceOpts: []git.TagSourceOptionFunc{
				git.TagSourceReplacing("nonexistent-", ""),
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tagsGetter, err := git.NewRepoTagsGetter(repodir, tc.tagOpts...)
			if err != nil {
				t.Fatalf("Error creating git source: %v", err)
			}

			src := git.NewTagsSource(tagsGetter, tc.tagSourceOpts...)

			hash, err := src.LastVersionHash()
			assert.ErrorIs(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedHash, hash, "Reported hash does not match")
		})
	}
}

func getVersionCommitHash(t *testing.T, repodir, version string, opts ...git.TagOptionFunc) string {
	t.Helper()

	tagsGetter, err := git.NewRepoTagsGetter(repodir, opts...)
	if err != nil {
		t.Fatalf("Error creating git source: %v", err)
	}

	tags, err := tagsGetter.Tags()
	if err != nil {
		t.Fatalf("Error fetching tags: %v", err)
	}

	for _, t := range tags {
		if t.Name == version {
			return t.Hash
		}
	}

	return ""
}
