package git_test

import (
	"errors"
	"testing"

	"github.com/newrelic/release-toolkit/src/git"
	"github.com/stretchr/testify/assert"
)

type fakeSource []git.Commit

func (fs fakeSource) Commits(_ string) ([]git.Commit, error) {
	return fs, nil
}

//nolint:funlen
func TestCommitFilter_Commits(t *testing.T) {
	t.Parallel()

	singleFileRoot := git.Commit{
		Files: []string{"single-file-on-root"},
	}
	singleFileFolder1 := git.Commit{
		Files: []string{"folder1/single-file-on-folder1"},
	}
	twoFilesFolder2 := git.Commit{
		Files: []string{"folder2/file1", "folder2/file2"},
	}
	threeFilesFolders := git.Commit{
		Files: []string{"folder1/file", "folder2/file", "folder3/file"},
	}
	rootAndFolder := git.Commit{
		Files: []string{"file-on-root", "folder1/file-on-folder"},
	}

	allCommits := []git.Commit{
		singleFileRoot, singleFileFolder1, twoFilesFolder2, threeFilesFolders, rootAndFolder,
	}

	for _, tc := range []struct {
		name            string
		opts            []git.CommitFilterOptionFunc
		commits         []git.Commit
		expectedCommits []git.Commit
	}{
		{
			name: "Included_Folders_All_Excluded_None",
			opts: []git.CommitFilterOptionFunc{},

			commits:         allCommits,
			expectedCommits: allCommits,
		},
		{
			name: "Include_Folder1",
			opts: []git.CommitFilterOptionFunc{
				git.IncludedDirs("folder1"),
			},
			commits: allCommits,
			expectedCommits: []git.Commit{
				singleFileFolder1, threeFilesFolders, rootAndFolder,
			},
		},
		{
			name: "Exclude_Folder1",
			opts: []git.CommitFilterOptionFunc{
				git.ExcludedDirs("folder1"),
			},
			commits: allCommits,
			expectedCommits: []git.Commit{
				singleFileRoot, twoFilesFolder2, threeFilesFolders, rootAndFolder,
			},
		},
		{
			name: "Exclude_Folder2",
			opts: []git.CommitFilterOptionFunc{
				git.ExcludedDirs("folder2"),
			},
			commits: allCommits,
			expectedCommits: []git.Commit{
				singleFileRoot, singleFileFolder1, threeFilesFolders, rootAndFolder,
			},
		},
		{
			name: "Include_Folder2",
			opts: []git.CommitFilterOptionFunc{
				git.IncludedDirs("folder2"),
			},
			commits: allCommits,
			expectedCommits: []git.Commit{
				twoFilesFolder2, threeFilesFolders,
			},
		},
		{
			name: "Include_Folder2_Exclude_Folder1",
			opts: []git.CommitFilterOptionFunc{
				git.IncludedDirs("folder2"),
				git.ExcludedDirs("folder1"),
			},
			commits: allCommits,
			expectedCommits: []git.Commit{
				// Commit will not be excluded as some changes scope.
				twoFilesFolder2, threeFilesFolders,
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			commitsFilter, err := git.NewCommitFilter(fakeSource(tc.commits), tc.opts...)
			if err != nil {
				t.Fatalf("Error creating git source: %v", err)
			}

			commits, err := commitsFilter.Commits("")
			if err != nil {
				t.Fatalf("Error fetching commits: %v", err)
			}

			assert.ElementsMatchf(t, tc.expectedCommits, commits, "Reported commits do not match")
		})
	}
}

func TestCommitFilter_NewCommitFilter(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		opts          []git.CommitFilterOptionFunc
		expectedError error
	}{
		{
			name:          "No_Included_Or_Excluded_Throws_No_Error",
			opts:          []git.CommitFilterOptionFunc{},
			expectedError: nil,
		},
		{
			name: "Leading_/_Symbol_Included_Throws_Error",
			opts: []git.CommitFilterOptionFunc{
				git.IncludedDirs("/"),
			},
			expectedError: git.ErrDirNotValid,
		},
		{
			name: "Leading_/_Symbol_Excluded_Throws_Error",
			opts: []git.CommitFilterOptionFunc{
				git.ExcludedDirs("/"),
			},
			expectedError: git.ErrDirNotValid,
		},
		{
			name: "Only_._Symbol_Included_Throws_Error",
			opts: []git.CommitFilterOptionFunc{
				git.IncludedDirs("."),
			},
			expectedError: git.ErrDirNotValid,
		},
		{
			name: "Only_.._Symbol_Included_Throws_Error",
			opts: []git.CommitFilterOptionFunc{
				git.IncludedDirs(".."),
			},
			expectedError: git.ErrDirNotValid,
		},
		{
			name: ".._Symbol_Excluded_Throws_Error",
			opts: []git.CommitFilterOptionFunc{
				git.ExcludedDirs("../a"),
			},
			expectedError: git.ErrDirNotValid,
		},
		{
			name: ".._In_Middle_Of_Path",
			opts: []git.CommitFilterOptionFunc{
				git.ExcludedDirs("a/../b"),
			},
			expectedError: git.ErrDirNotValid,
		},
		{
			name: "._In_Middle_Of_Path",
			opts: []git.CommitFilterOptionFunc{
				git.ExcludedDirs("a/./b"),
			},
			expectedError: git.ErrDirNotValid,
		},
		{
			name: "Correct_Paths_Throw_No_Error",
			opts: []git.CommitFilterOptionFunc{
				git.ExcludedDirs(".a/b.c", "a"),
			},
			expectedError: nil,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := git.NewCommitFilter(git.NewRepoCommitsGetter("a-repo-path"), tc.opts...)
			if !errors.Is(err, tc.expectedError) {
				t.Fatalf("Error expected: %v but was: %v", tc.expectedError, err)
			}
		})
	}
}
