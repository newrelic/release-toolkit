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

type SemverTags struct {
	Versions []*semver.Version
	Hashes   map[*semver.Version]string
}

type SemverTagsGetter interface {
	Tags() (SemverTags, error)
}

type RepoSemverTagsGetter struct {
	workDir  string
	match    *regexp.Regexp
	replacer *strings.Replacer
}

type TagOptionFunc func(s *RepoSemverTagsGetter) error

// TagsMatching returns an option that will make the getter to ignore tags that do not match regex.
func TagsMatching(regex string) TagOptionFunc {
	return func(s *RepoSemverTagsGetter) error {
		rgx, err := regexp.Compile(regex)
		if err != nil {
			return fmt.Errorf("compiling %q: %w", regex, err)
		}

		s.match = rgx
		return nil
	}
}

// TagsReplacing returns an option that will perform a string replacement on tags
// that match the regex before attempting to parse them as versions.
// It is useful to, for example, strip prefixes matched with TagMatching.
func TagsReplacing(existing, replacement string) TagOptionFunc {
	return func(s *RepoSemverTagsGetter) error {
		s.replacer = strings.NewReplacer(existing, replacement)
		return nil
	}
}

var MatchAllTags = regexp.MustCompile("")

func NewRepoSemverTagsGetter(workDir string, opts ...TagOptionFunc) (*RepoSemverTagsGetter, error) {
	s := &RepoSemverTagsGetter{
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

func (s *RepoSemverTagsGetter) Tags() (SemverTags, error) {
	repo, err := git.PlainOpen(s.workDir)
	if err != nil {
		return SemverTags{}, fmt.Errorf("opening git repo at %s: %w", s.workDir, err)
	}

	repoTags, err := repo.Tags()
	if err != nil {
		return SemverTags{}, fmt.Errorf("getting git tags: %w", err)
	}

	tags := SemverTags{
		Hashes: make(map[*semver.Version]string),
	}

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

		tags.Versions = append(tags.Versions, v)
		tags.Hashes[v] = reference.Hash().String()

		return nil
	})
	if err != nil {
		return SemverTags{}, fmt.Errorf("iterating over tags: %w", err)
	}

	return tags, nil
}
