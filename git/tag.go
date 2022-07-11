package git

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"github.com/newrelic/release-toolkit/tag"
)

// TagSource allows tags from a git repository as a source for previous versions.
type TagSource struct {
	workDir  string
	match    *regexp.Regexp
	replacer *strings.Replacer
}

type TagOptionFunc func(s *TagSource) error

// TagMatching returns an option that will cause Source to ignore tags or commits that do not match regex.
func TagMatching(regex string) TagOptionFunc {
	return func(s *TagSource) error {
		rgx, err := regexp.Compile(regex)
		if err != nil {
			return fmt.Errorf("compiling %q: %w", regex, err)
		}

		s.match = rgx
		return nil
	}
}

// TagReplacing returns an option that will replace a string in the tag source.
func TagReplacing(old, new string) TagOptionFunc {
	return func(s *TagSource) error {
		s.replacer = strings.NewReplacer(old, new)
		return nil
	}
}

var MatchAllTags = regexp.MustCompile("")

func NewTagSource(workDir string, opts ...TagOptionFunc) (*TagSource, error) {
	s := &TagSource{
		workDir:  workDir,
		match:    MatchAllTags,
		replacer: strings.NewReplacer(),
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("applyng option: %w", err)
		}
	}

	return s, nil
}

func (s *TagSource) Tags() ([]tag.Tag, error) {
	repo, err := git.PlainOpen(s.workDir)
	if err != nil {
		return nil, fmt.Errorf("opening git repo at %s: %w", s.workDir, err)
	}

	repoTags, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("getting git tags: %w", err)
	}

	var tags []tag.Tag

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

		tagName = s.replacer.Replace(tagName)

		v, innerErr := semver.NewVersion(tagName)
		if innerErr != nil {
			log.Debugf("skipping tag %q as it does not conform to semver %v", tagName, innerErr)
			return nil
		}

		tags = append(tags, tag.Tag{
			Version: v,
			Hash:    reference.Hash().String(),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("iterating over tags: %w", err)
	}

	return tags, nil
}
