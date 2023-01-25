package git

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// EmptyTreeID is the universal git empty tree sha1.
const EmptyTreeID = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

type Commit struct {
	Message string
	Hash    string
	Author  string
	Files   []string
}

type CommitsGetter interface {
	Commits(lastHash string) ([]Commit, error)
}

// RepoCommitsGetter gets commits from a git repository.
type RepoCommitsGetter struct {
	workDir string
}

func NewRepoCommitsGetter(workDir string) *RepoCommitsGetter {
	return &RepoCommitsGetter{
		workDir: workDir,
	}
}

// Commits returns all the commits from Head ordered from top to bottom
// until LastHash, if lastHash is empty, all commits are returned.
func (s *RepoCommitsGetter) Commits(lastHash string) ([]Commit, error) {
	repo, err := git.PlainOpen(s.workDir)
	if err != nil {
		return nil, fmt.Errorf("opening git repo at %s: %w", s.workDir, err)
	}

	commitIter, err := repo.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
	if err != nil {
		return nil, fmt.Errorf("getting git commits: %w", err)
	}
	defer commitIter.Close()

	var isLastHash bool
	gitCommits := make([]*object.Commit, 0)

	for !isLastHash {
		cm, errCommit := commitIter.Next()
		if errCommit != nil && !errors.Is(errCommit, io.EOF) {
			return nil, fmt.Errorf("iterating git commits: %w", errCommit)
		}

		// No more commits to iterate
		if cm == nil {
			break
		}

		gitCommits = append(gitCommits, cm)

		// Get also LastHash to be able to compare it with last in list
		if cm.Hash.String() == lastHash {
			isLastHash = true
		}
	}

	return s.commitsWithChangedFiles(lastHash, gitCommits)
}

// commitsWithChanges iterates the list of commits and populates a list with changed files.
func (s *RepoCommitsGetter) commitsWithChangedFiles(lastHash string, gitCommits []*object.Commit) ([]Commit, error) {
	commits := make([]Commit, 0)

	for k, cm := range gitCommits {
		currentTree, err := cm.Tree()
		if err != nil {
			return nil, fmt.Errorf("getting commit tree: %w", err)
		}

		prevTree, err := s.getPreviousCommitTree(k, gitCommits)
		if err != nil {
			return nil, fmt.Errorf("getting previous commit tree: %w", err)
		}

		// Get changes between current commit and previous
		changes, err := prevTree.Diff(currentTree)
		if err != nil {
			return nil, fmt.Errorf("getting diff between commits: %w", err)
		}

		if cm.Hash.String() == lastHash {
			continue
		}

		commits = append(commits, Commit{
			Message: strings.TrimSuffix(cm.Message, "\n"),
			Hash:    cm.Hash.String(),
			Author:  cm.Author.String(),
			Files:   s.getChangedFiles(changes),
		})
	}

	return commits, nil
}

//nolint:wrapcheck
func (s *RepoCommitsGetter) getPreviousCommitTree(positionInList int, commitList []*object.Commit) (*object.Tree, error) {
	if positionInList < len(commitList)-1 {
		return commitList[positionInList+1].Tree()
	}

	// When no previous commit exits, we get the empty tree
	return &object.Tree{
		Hash: plumbing.NewHash(EmptyTreeID),
	}, nil
}

func (s *RepoCommitsGetter) getChangedFiles(changes []*object.Change) []string {
	changedFiles := make([]string, 0)
	for _, change := range changes {
		empty := object.ChangeEntry{}
		if change.From != empty {
			changedFiles = append(changedFiles, change.From.Name)
			continue
		}
		changedFiles = append(changedFiles, change.To.Name)
	}

	return changedFiles
}
