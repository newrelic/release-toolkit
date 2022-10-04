package git_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/newrelic/release-toolkit/src/git"
	"github.com/stretchr/testify/assert"
)

type testCommitTag struct {
	msgTag string
	files  []string
}

func repoWithCommitsAndTags(t *testing.T, commitsAndTags ...testCommitTag) string {
	t.Helper()

	dir := t.TempDir()

	cmds := []string{
		"git init",
		"git config user.email test@user.tld",
		"git config user.name Test",
		"git config commit.gpgsign false",
	}

	// Create empty files, commit and tag it for each msgTag name.
	for _, ct := range commitsAndTags {
		for _, file := range ct.files {
			path := filepath.Dir(file)
			if path != "." {
				if err := os.Mkdir(filepath.Join(dir, path), os.ModePerm); err != nil && os.IsNotExist(err) {
					t.Fatalf("Error creating dir: %v", err)
				}
			}
			cmds = append(cmds, fmt.Sprintf("touch %s", file))
			cmds = append(cmds, fmt.Sprintf("git add %s", file))
		}

		cmds = append(cmds, fmt.Sprintf("git commit -m %s", ct.msgTag))
		cmds = append(cmds, fmt.Sprintf("git tag %s", ct.msgTag))
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
	t.Parallel()
	repodir := repoWithCommitsAndTags(t,
		testCommitTag{"v1.2.3", []string{"v1.2.3"}},
		testCommitTag{"v1.3.0", []string{"v1.3.0"}},
		testCommitTag{"v1.4.0", []string{"v1.4.0"}},
		testCommitTag{"1.5.0", []string{"1.5.0"}},
		testCommitTag{"2.0.0-beta", []string{"2.0.0-beta"}},
	)

	for _, tc := range []struct {
		name            string
		commitHash      string
		expectedCommits []string
	}{
		{
			name: "Empty_Hash_Default_Opts",
			expectedCommits: []string{
				"2.0.0-beta",
				"1.5.0",
				"v1.4.0",
				"v1.3.0",
				"v1.2.3",
			},
		},
		{
			name:       "Non_Existing_Hash_Default_Opts",
			commitHash: "an-invented-hash",
			expectedCommits: []string{
				"2.0.0-beta",
				"1.5.0",
				"v1.4.0",
				"v1.3.0",
				"v1.2.3",
			},
		},
		{
			name:       "Existing_Hash_Default_Opts",
			commitHash: getVersionCommitHash(t, repodir, "v1.4.0"),
			expectedCommits: []string{
				"2.0.0-beta",
				"1.5.0",
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			commitsGetter := git.NewRepoCommitsGetter(repodir)
			commits, err := commitsGetter.Commits(tc.commitHash)
			if err != nil {
				t.Fatalf("Error fetching commits: %v", err)
			}

			strCommits := make([]string, 0, len(commits))
			for _, c := range commits {
				strCommits = append(strCommits, c.Message)
			}

			assert.ElementsMatchf(t, tc.expectedCommits, strCommits, "Reported commits do not match")
		})
	}
}
