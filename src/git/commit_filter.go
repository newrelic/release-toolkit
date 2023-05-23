package git

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var (
	ErrDirNotValid  = errors.New(`invalid dir path, it can't be ".", "..", start by "/" or contain "./"`)
	ErrFileNotValid = errors.New(`invalid file path, it can't be ".", "..", start or end by "/" or contain "./"`)
)

// CommitFilter filters commits from a git repository based in included and excluded directories.
type CommitFilter struct {
	commitsGetter CommitsGetter
	includedDirs  []string
	excludedDirs  []string
	includedFiles []string
	excludedFiles []string
}

type CommitFilterOptionFunc func(s *CommitFilter) error

// IncludedDirs returns an option that will filter commits with all changes not in includedDirs.
func IncludedDirs(includedDirs ...string) CommitFilterOptionFunc {
	return func(s *CommitFilter) error {
		for _, dir := range includedDirs {
			if isDirNameInvalid(dir) {
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
			if isDirNameInvalid(dir) {
				return fmt.Errorf("excluded dir %s : %w", dir, ErrDirNotValid)
			}
		}
		s.excludedDirs = excludedDirs
		return nil
	}
}

// IncludedFiles returns an option that will filter commits with all changes not in includedFiles.
func IncludedFiles(includedFiles ...string) CommitFilterOptionFunc {
	return func(s *CommitFilter) error {
		for _, file := range includedFiles {
			if isFileNameInvalid(file) {
				return fmt.Errorf("included file %s : %w", file, ErrFileNotValid)
			}
		}
		s.includedFiles = includedFiles
		return nil
	}
}

// ExcludedFiles returns an option that will filter commits with all changes in excludedFiles.
func ExcludedFiles(excludedFiles ...string) CommitFilterOptionFunc {
	return func(s *CommitFilter) error {
		for _, file := range excludedFiles {
			if isFileNameInvalid(file) {
				return fmt.Errorf("excluded file %s : %w", file, ErrFileNotValid)
			}
		}
		s.excludedFiles = excludedFiles
		return nil
	}
}

func isDirNameInvalid(dir string) bool {
	return dir == "." || dir == ".." || strings.HasPrefix(dir, "/") || strings.Contains(dir, "./")
}

func isFileNameInvalid(file string) bool {
	// checks are the very same for directories, adding the check the trailing '/'
	return strings.HasSuffix(file, "/") || isDirNameInvalid(file)
}

func NewCommitFilter(commitsGetter CommitsGetter, opts ...CommitFilterOptionFunc) (*CommitFilter, error) {
	cf := &CommitFilter{
		commitsGetter: commitsGetter,
	}

	for _, opt := range opts {
		if err := opt(cf); err != nil {
			return nil, fmt.Errorf("commit filter, applying option: %w", err)
		}
	}

	return cf, nil
}

// Commits calls commitGetter to get a list of commits until lastHash.
// If includedDirs or includedFiles is not empty, only commits changing at least one file contained in any of the includedDirs or includedFiles will be returned.
// If excludedDirs or excludedFiles is not empty, commits that only change files present in excludedDirs or excludedFiles will be filtered out.
// If you specify both include and exclude filters, the commits are filtered out.
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

// commitChangesIncluded returns true if at least one of the changes is included in includedDirs or includedFiles.
func (s *CommitFilter) commitChangesIncluded(files []string) bool {
	if len(s.includedDirs) < 1 && len(s.includedFiles) < 1 {
		return true
	}

	for _, file := range files {
		for _, includedDir := range s.includedDirs {
			if strings.HasPrefix(filepath.Dir(file)+"/", filepath.Clean(includedDir)+"/") {
				return true
			}
		}

		for _, includedFile := range s.includedFiles {
			if file == includedFile {
				return true
			}
		}
	}

	return false
}

// commitChangesExcluded returns true if all changes are in excludedDirs or in excludedFiles.
func (s *CommitFilter) commitChangesExcluded(files []string) bool {
	if len(s.excludedDirs) < 1 && len(s.excludedFiles) < 1 {
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

		for _, excludedFile := range s.excludedFiles {
			if file == excludedFile {
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
