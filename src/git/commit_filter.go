package git

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var (
	ErrDirNotValid  = errors.New(`invalid dir path, it can't be "",".", "..", start by "/" or contain "./"`)
	ErrFileNotValid = errors.New(`invalid file path, it can't be "",".", "..", start or end by "/" or contain "./"`)
)

// CommitFilter filters commits from a git repository based in included and excluded directories.
type CommitFilter struct {
	commitsGetter           CommitsGetter
	includedDirs            []string
	excludedDirs            []string
	includedFiles           []string
	excludedFiles           []string
	excludedDevDependencies []string
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

func ExcludedDevDependencies(deps ...string) CommitFilterOptionFunc {
	return func(s *CommitFilter) error {
		s.excludedDevDependencies = deps
		return nil
	}
}

func isDirNameInvalid(dir string) bool {
	return dir == "." || dir == ".." || dir == "" || strings.HasPrefix(dir, "/") || strings.Contains(dir, "./")
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
// If includedDirs or includedFiles is not empty, commits changing only files not contained in any of the includedDirs or includedFiles will be filtered out.
// Moreover, commits modifying only files present in excludedDirs or excludedFiles will be filtered out.
// Notice that if a file is excluded by a rule it is filtered out, even if another rule include it.
func (s *CommitFilter) Commits(lastHash string) ([]Commit, error) {
	commits, err := s.commitsGetter.Commits(lastHash)
	if err != nil {
		return nil, fmt.Errorf("commit filter, getting commits: %w", err)
	}

	filteredCommits := make([]Commit, 0)
	for _, commit := range commits {
		if !s.commitChangesExcluded(commit.Files) && !s.commitDevDependenciesExcluded(commit.Message) {
			filteredCommits = append(filteredCommits, commit)
		}
	}

	return filteredCommits, nil
}

// commitChangesExcluded returns true if all changes are excluded or if none of the changes are included.
// Notice that the exclude-clause takes precedence.
func (s *CommitFilter) commitChangesExcluded(files []string) bool {
	for _, file := range files {
		var changeIsExcluded bool

		if !s.fileIncluded(file) {
			changeIsExcluded = true
		}

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

// commitDevDependenciesExcluded checks if the commit message contains any excluded dev dependencies.
func (s *CommitFilter) commitDevDependenciesExcluded(message string) bool {
	for _, dep := range s.excludedDevDependencies {
		if strings.Contains(message, dep) {
			return true
		}
	}

	return false
}

// commitChangesIncluded returns true if the file is included in includedDirs or includedFiles.
func (s *CommitFilter) fileIncluded(file string) bool {
	if len(s.includedDirs) < 1 && len(s.includedFiles) < 1 {
		return true
	}

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

	return false
}
