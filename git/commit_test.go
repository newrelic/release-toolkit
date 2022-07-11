package git_test

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/newrelic/release-toolkit/git"
	"github.com/stretchr/testify/assert"
)

func repoWithCommitsAndTags(t *testing.T, commitsAndTags ...string) string {
	t.Helper()

	dir := t.TempDir()

	cmds := []string{
		"git init",
		"git config user.email test@user.tld",
		"git config user.name Test",
	}

	for _, t := range commitsAndTags {
		cmds = append(cmds, fmt.Sprintf("touch %s", t))
		cmds = append(cmds, fmt.Sprintf("git add %s", t))
		cmds = append(cmds, fmt.Sprintf("git commit -m %s", t))
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

func TestCommitSource_Commits(t *testing.T) {
	repodir := repoWithCommitsAndTags(t,
		"v1.2.3",
		"v1.3.0",
		"v1.4.0",
		"1.5.0",
		"2.0.0-beta",
	)

	tagSrc, err := git.NewTagSource(repodir)
	if err != nil {
		t.Fatalf("Error creating git source: %v", err)
	}

	tags, err := tagSrc.Tags()
	if err != nil {
		t.Fatalf("Error fetching tags: %v", err)
	}

	for _, tc := range []struct {
		name            string
		commitHash      string
		opts            []git.CommitOptionFunc
		expectedCommits []string
	}{
		{
			name: "Empty_Hash_Default_Opts",
			expectedCommits: []string{
				"2.0.0-beta\n",
				"1.5.0\n",
				"v1.4.0\n",
				"v1.3.0\n",
				"v1.2.3\n",
			},
		},
		{
			name: "Empty_Hash_Matching_Leading_v",
			opts: []git.CommitOptionFunc{git.CommitMessageMatching("^v")},
			expectedCommits: []string{
				"v1.4.0\n",
				"v1.3.0\n",
				"v1.2.3\n",
			},
		},
		{
			name: "Empty_Hash_Matching_Author",
			opts: []git.CommitOptionFunc{
				git.CommitAuthorMatching("unknown-author"),
			},
			expectedCommits: nil,
		},
		{
			name: "Empty_Hash_Not_Matching_Author",
			opts: []git.CommitOptionFunc{
				git.CommitAuthorMatching("Test"),
			},
			expectedCommits: []string{
				"2.0.0-beta\n",
				"1.5.0\n",
				"v1.4.0\n",
				"v1.3.0\n",
				"v1.2.3\n",
			},
		},
		{
			name:       "Non_Existing_Hash_Default_Opts",
			commitHash: "an-invented-hash",
			expectedCommits: []string{
				"2.0.0-beta\n",
				"1.5.0\n",
				"v1.4.0\n",
				"v1.3.0\n",
				"v1.2.3\n",
			},
		},
		{
			name:       "Existing_Hash_Default_Opts",
			commitHash: tags[2].Hash,
			expectedCommits: []string{
				"2.0.0-beta\n",
				"1.5.0\n",
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			src, err := git.NewCommitSource(repodir, tc.opts...)
			if err != nil {
				t.Fatalf("Error creating git source: %v", err)
			}

			commits, err := src.Commits(tc.commitHash)
			if err != nil {
				t.Fatalf("Error fetching commits: %v", err)
			}

			strCommits := make([]string, 0, len(tags))
			for _, c := range commits {
				strCommits = append(strCommits, c.Message)
			}

			assert.ElementsMatchf(t, tc.expectedCommits, strCommits, "Reported commits do not match")
		})
	}
}
