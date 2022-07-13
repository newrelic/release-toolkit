package git

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
)

type TagsReleaseGetter interface {
	Versions() ([]*semver.Version, error)
	LastReleaseHash() (string, error)
}

// TagsSource implements the `version.Source` interface, using tags from a git repository as a source for previous versions.
// It also implements TagsReleaseGetter to be used by extractor services.
type TagsSource struct {
	tagsGetter TagsGetter
	replacer   *strings.Replacer
}

type TagSourceOptionFunc func(s *TagsSource)

// TagSourceReplacing returns an option that will perform a string replacement on tags
// that match the regex before attempting to parse them as versions.
// It is useful to, for example, strip prefixes from tags matched with TagMatching.
func TagSourceReplacing(existing, replacement string) TagSourceOptionFunc {
	return func(s *TagsSource) {
		s.replacer = strings.NewReplacer(existing, replacement)
	}
}

func NewTagsSource(tagsGetter TagsGetter, opts ...TagSourceOptionFunc) *TagsSource {
	ts := &TagsSource{
		tagsGetter: tagsGetter,
		replacer:   strings.NewReplacer(),
	}

	for _, opt := range opts {
		opt(ts)
	}

	return ts
}

func (s *TagsSource) Versions() ([]*semver.Version, error) {
	tags, err := s.tagsGetter.Tags()
	if err != nil {
		return nil, fmt.Errorf("getting tags: %w", err)
	}

	versions := make([]*semver.Version, 0)
	for _, tag := range tags {
		tagName := s.replacer.Replace(tag.Name)

		v, innerErr := semver.NewVersion(tagName)
		if innerErr != nil {
			log.Infof("skipping tag %q as it does not conform to semver %v", tagName, innerErr)
			continue
		}

		versions = append(versions, v)
	}

	return versions, nil
}

func (s *TagsSource) LastReleaseHash() (string, error) {
	tags, err := s.tagsGetter.Tags()
	if err != nil {
		return "", fmt.Errorf("getting tags: %w", err)
	}

	sort.Slice(tags, func(i, j int) bool {
		tagNameCurrent := s.replacer.Replace(tags[i].Name)
		vCurrent, innerErr := semver.NewVersion(tags[i].Name)
		if innerErr != nil {
			log.Debugf("skipping tag %q as it does not conform to semver %v", tagNameCurrent, innerErr)
			return false
		}

		tagNameNext := s.replacer.Replace(tags[j].Name)
		vNext, innerErr := semver.NewVersion(tags[j].Name)
		if innerErr != nil {
			log.Debugf("skipping tag %q as it does not conform to semver %v", tagNameNext, innerErr)
			return true
		}
		return vCurrent.GreaterThan(vNext)
	})

	return tags[0].Hash, nil
}
