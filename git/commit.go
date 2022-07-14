package git

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
)

type Commit struct {
	Message string
	Hash    string
	Author  string
}

type CommitsGetter interface {
	Commits(lastHash string) ([]Commit, error)
}

// RepoCommitsGetter gets commits from a git repository.
type RepoCommitsGetter struct {
	workDir      string
	matchMessage *regexp.Regexp
	matchAuthor  *regexp.Regexp
}

type CommitOptionFunc func(s *RepoCommitsGetter) error

// CommitMessageMatching returns an option that will cause Source to ignore commits with Message not matching regex.
func CommitMessageMatching(regex string) CommitOptionFunc {
	return func(s *RepoCommitsGetter) error {
		rgx, err := regexp.Compile(regex)
		if err != nil {
			return fmt.Errorf("compiling %q: %w", regex, err)
		}

		s.matchMessage = rgx
		return nil
	}
}

// CommitAuthorMatching returns an option that will cause Source to ignore commits with Author not matching regex.
func CommitAuthorMatching(regex string) CommitOptionFunc {
	return func(s *RepoCommitsGetter) error {
		rgx, err := regexp.Compile(regex)
		if err != nil {
			return fmt.Errorf("compiling %q: %w", regex, err)
		}

		s.matchAuthor = rgx
		return nil
	}
}

var MatchAllCommits = regexp.MustCompile("")

func NewRepoCommitsGetter(workDir string, opts ...CommitOptionFunc) (*RepoCommitsGetter, error) {
	s := &RepoCommitsGetter{
		workDir:      workDir,
		matchMessage: MatchAllCommits,
		matchAuthor:  MatchAllCommits,
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("applyng option: %w", err)
		}
	}

	return s, nil
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

	var commits []Commit

	for cm, errCommit := commitIter.Next(); errCommit == nil; {
		if cm.Hash.String() == lastHash {
			break
		}

		if !s.matchMessage.MatchString(cm.Message) {
			log.Debugf("skipping commit %q as message does not match %q", cm.Message, s.matchMessage.String())
			cm, errCommit = commitIter.Next()
			continue
		}

		if !s.matchAuthor.MatchString(cm.Author.Name) {
			log.Debugf("skipping commit %q as it author does not match %q", cm.Author, s.matchAuthor.String())
			cm, errCommit = commitIter.Next()
			continue
		}

		commits = append(commits, Commit{
			Message: strings.TrimSuffix(cm.Message, "\n"),
			Hash:    cm.Hash.String(),
		})
		cm, errCommit = commitIter.Next()
	}

	commitIter.Close()

	return commits, nil
}
