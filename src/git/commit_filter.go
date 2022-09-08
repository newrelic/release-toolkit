package git

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var ErrDirNotValid = errors.New(`invalid dir path, it can't be ".", "..", start by "/" or contain "./"`)

// CommitFilter filters commits from a git repository based in included and excluded directories.
type CommitFilter struct {
	commitsGetter CommitsGetter
	includedDirs  []string
	excludedDirs  []string
}

type CommitFilterOptionFunc func(s *CommitFilter) error

// IncludedDirs returns an option that will filter commits with all changes not in includedDirs.
func IncludedDirs(includedDirs ...string) CommitFilterOptionFunc {
	return func(s *CommitFilter) error {
		for _, dir := range includedDirs {
			if dir == "." || dir == ".." || strings.HasPrefix(dir, "/") || strings.Contains(dir, "./") {
				return fmt.Errorf("included dir %s : %w", dir, ErrDirNotValid)
			}
		}
		s.includedDirs = includedDirs
		return nil
	}
}

// ExcludedDirs returns an option that will filter commits with all changes in excludedDirs.
func ExcludedDirs(excludedDirs ...string) CommitFilterOptionFunc {
	return func(s *CommitFilter) error {
		for _, dir := range excludedDirs {
			if dir == "." || dir == ".." || strings.HasPrefix(dir, "/") || strings.Contains(dir, "./") {
				return fmt.Errorf("excluded dir %s : %w", dir, ErrDirNotValid)
			}
		}
		s.excludedDirs = excludedDirs
		return nil
	}
}

func NewCommitFilter(commitsGetter CommitsGetter, opts ...CommitFilterOptionFunc) (*CommitFilter, error) {
	cf := &CommitFilter{
		commitsGetter: commitsGetter,
	}

	for _, opt := range opts {
		if err := opt(cf); err != nil {
			return nil, fmt.Errorf("commit filter, applyng option: %w", err)
		}
	}

	return cf, nil
}

// Commits calls commitGetter to get a list of commits until lastHash.
// If includedDirs is not empty, only commits changing at least one file contained in any of the includedDirs will be returned.
// If excludedDirs is not empty, commits that only change files present in excludedDirs will be filtered out.
func (s *CommitFilter) Commits(lastHash string) ([]Commit, error) {
	commits, err := s.commitsGetter.Commits(lastHash)
	if err != nil {
		return nil, fmt.Errorf("commit filter, getting commits: %w", err)
	}

	filteredCommits := make([]Commit, 0)
	for _, commit := range commits {
		if s.commitChangesIncluded(commit.Files) && !s.commitChangesExcluded(commit.Files) {
			filteredCommits = append(filteredCommits, commit)
		}
	}

	return filteredCommits, nil
}

// commitChangesIncluded returns true if at least one of the changes is included in includedDirs.
func (s *CommitFilter) commitChangesIncluded(files []string) bool {
	if len(s.includedDirs) < 1 {
		return true
	}

	for _, file := range files {
		for _, includedDir := range s.includedDirs {
			if strings.HasPrefix(filepath.Dir(file)+"/", filepath.Clean(includedDir)+"/") {
				return true
			}
		}
	}

	return false
}

// commitChangesExcluded returns true if all changes are in excludedDirs.
func (s *CommitFilter) commitChangesExcluded(files []string) bool {
	if len(s.excludedDirs) < 1 {
		return false
	}

	for _, file := range files {
		var changeIsExcluded bool
		for _, excludedDir := range s.excludedDirs {
			if strings.HasPrefix(filepath.Dir(file)+"/", filepath.Clean(excludedDir)+"/") {
				changeIsExcluded = true
				break
			}
		}

		if !changeIsExcluded {
			return false
		}
	}

	return true
}
