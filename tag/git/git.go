package git

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Source allows tags from a git repository as a source for previous versions.
type Source struct {
	workDir  string
	match    *regexp.Regexp
	replacer *strings.Replacer
}

type OptionFunc func(s *Source) error

// Matching returns an option that will cause Source to ignore tags that do not match regex.
func Matching(regex string) OptionFunc {
	return func(s *Source) error {
		rgx, err := regexp.Compile(regex)
		if err != nil {
			return fmt.Errorf("compiling %q: %w", regex, err)
		}

		s.match = rgx
		return nil
	}
}

func Replacing(existing, replacement string) OptionFunc {
	return func(s *Source) error {
		s.replacer = strings.NewReplacer(existing, replacement)
		return nil
	}
}

var MatchAll = regexp.MustCompile("")

func NewSource(workDir string, opts ...OptionFunc) (*Source, error) {
	s := &Source{
		workDir:  workDir,
		match:    MatchAll,
		replacer: strings.NewReplacer(),
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("applyng option: %w", err)
		}
	}

	return s, nil
}

func (s *Source) Tags() ([]*semver.Version, error) {
	repo, err := git.PlainOpen(s.workDir)
	if err != nil {
		return nil, fmt.Errorf("opening git repo at %s: %w", s.workDir, err)
	}

	tags, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("getting git tags: %w", err)
	}

	var versions []*semver.Version

	err = tags.ForEach(func(reference *plumbing.Reference) error {
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

		versions = append(versions, v)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("iterating over tags: %w", err)
	}

	return versions, nil
}
