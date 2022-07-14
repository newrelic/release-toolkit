package git_test

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/newrelic/release-toolkit/git"
	"github.com/stretchr/testify/assert"
)

func repoWithTags(t *testing.T, tags ...string) string {
	t.Helper()

	dir := t.TempDir()

	cmds := []string{
		"git init",
		"git config user.email test@user.tld",
		"git config user.name Test",
		"touch a",
		"git add a",
		"git commit -m test",
	}

	for _, t := range tags {
		cmds = append(cmds, fmt.Sprintf("git tag %s", t))
	}

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

	return dir
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
			tagOpts: []git.TagOptionFunc{git.TagsMatching("^v")},
			expectedTags: []string{
				"1.4.0",
				"1.3.0",
				"1.2.3",
			},
		},
		{
			name: "Matching_And_Replacing_Prefix",
			tagOpts: []git.TagOptionFunc{
				git.TagsMatching("^helm-chart-"),
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
