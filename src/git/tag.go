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
	workDir      string
	matchRegex   *regexp.Regexp
	matchCommits map[string]bool
}

type TagOptionFunc func(s *RepoTagsGetter) error

// TagsMatchingRegex returns an option that will make the getter to ignore tags that do not match regex.
func TagsMatchingRegex(regex string) TagOptionFunc {
	return func(s *RepoTagsGetter) error {
		rgx, err := regexp.Compile(regex)
		if err != nil {
			return fmt.Errorf("compiling %q: %w", regex, err)
		}

		s.matchRegex = rgx
		return nil
	}
}

// TagsMatchingCommits returns an option that will make the getter to ignore tags that do not match commits.
func TagsMatchingCommits(getter CommitsGetter) TagOptionFunc {
	return func(s *RepoTagsGetter) error {
		branchCommits, err := getter.Commits(EmptyTreeID)
		if err != nil {
			return fmt.Errorf("getting git commits since empty tree: %w", err)
		}

		s.matchCommits = map[string]bool{}
		for _, c := range branchCommits {
			s.matchCommits[c.Hash] = true
		}

		return nil
	}
}

var MatchAllTags = regexp.MustCompile("")

func NewRepoTagsGetter(workDir string, opts ...TagOptionFunc) (*RepoTagsGetter, error) {
	s := &RepoTagsGetter{
		workDir:    workDir,
		matchRegex: MatchAllTags,
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("applying option: %w", err)
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

	var tags []Tag

	err = repoTags.ForEach(func(reference *plumbing.Reference) error {
		ref := reference.Name()
		if !ref.IsTag() {
			log.Tracef("Ignoring reference %q, it is not a tag", ref.String())
			return nil
		}

		tagName := ref.Short()
		if !s.matchRegex.MatchString(tagName) {
			log.Debugf("skipping tag %q as it does not match %q", tagName, s.matchRegex.String())
			return nil
		}

		if s.matchCommits != nil {
			if _, ok := s.matchCommits[reference.Hash().String()]; !ok {
				log.Infof("Ignoring %s since it belongs to a different branch", reference.Name().Short())
				return nil
			}
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
