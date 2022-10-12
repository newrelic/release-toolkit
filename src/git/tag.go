package git

import (
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type Tag struct {
	Name string
	Hash string
}

type TagsGetter interface {
	Tags() ([]Tag, error)
}

type RepoTagsGetter struct {
	workDir       string
	match         *regexp.Regexp
	commitsGetter CommitsGetter
}

type TagOptionFunc func(s *RepoTagsGetter) error

// TagsMatching returns an option that will make the getter to ignore tags that do not match regex.
func TagsMatching(regex string) TagOptionFunc {
	return func(s *RepoTagsGetter) error {
		rgx, err := regexp.Compile(regex)
		if err != nil {
			return fmt.Errorf("compiling %q: %w", regex, err)
		}

		s.match = rgx
		return nil
	}
}

var MatchAllTags = regexp.MustCompile("")

func NewRepoTagsGetter(workDir string, getter CommitsGetter, opts ...TagOptionFunc) (*RepoTagsGetter, error) {
	s := &RepoTagsGetter{
		workDir:       workDir,
		match:         MatchAllTags,
		commitsGetter: getter,
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("applyng option: %w", err)
		}
	}

	return s, nil
}

func (s *RepoTagsGetter) Tags() ([]Tag, error) {
	repo, err := git.PlainOpen(s.workDir)
	if err != nil {
		return nil, fmt.Errorf("opening git repo at %s: %w", s.workDir, err)
	}

	repoTags, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("getting git tags: %w", err)
	}

	branchCommits, err := getCommitsMap(s.commitsGetter)
	if err != nil {
		return nil, fmt.Errorf("getting commit map to filter tags: %w", err)
	}

	var tags []Tag

	err = repoTags.ForEach(func(reference *plumbing.Reference) error {
		ref := reference.Name()
		if !ref.IsTag() {
			log.Tracef("Ignoring reference %q, it is not a tag", ref.String())
			return nil
		}

		tagName := ref.Short()
		if !s.match.MatchString(tagName) {
			log.Debugf("skipping tag %q as it does not match %q", tagName, s.match.String())
			return nil
		}

		if _, ok := branchCommits[reference.Hash().String()]; !ok {
			log.Infof("Ignoring %s since it belongs to a different branch", reference.Name().Short())
			return nil
		}

		tags = append(tags, Tag{
			Name: tagName,
			Hash: reference.Hash().String(),
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("iterating over tags: %w", err)
	}

	return tags, nil
}

// getCommitsMap retrieves the list of commits from HEAD and put it in a map in order to
// filter out all the tags not belonging to the current branch
func getCommitsMap(commitGetter CommitsGetter) (map[string]bool, error) {
	branchCommits, err := commitGetter.Commits(EmptyTreeID)
	if err != nil {
		return nil, fmt.Errorf("getting git commits since empty tree: %w", err)
	}

	commitMap := map[string]bool{}
	for _, c := range branchCommits {
		commitMap[c.Hash] = true
	}

	return commitMap, nil
}
