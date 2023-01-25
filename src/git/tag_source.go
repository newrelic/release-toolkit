package git

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
)

var ErrNoReleases = errors.New("no releases found")

type TagsVersionGetter interface {
	Versions() ([]*semver.Version, error)
	LastVersionHash() (string, error)
}

// TagsSource implements the `version.Source` interface, using tags from a git repository as a source for previous versions.
// It also implements TagsVersionGetter to be used by extractor services.
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

	versions := make([]*semver.Version, 0, len(tags))
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

func (s *TagsSource) LastVersionHash() (string, error) {
	tags, err := s.tagsGetter.Tags()
	if err != nil {
		return "", fmt.Errorf("getting tags: %w", err)
	}

	type semverTag struct {
		tag     Tag
		version *semver.Version
	}

	versions := make([]semverTag, 0, len(tags))
	for _, tag := range tags {
		tagName := s.replacer.Replace(tag.Name)

		v, innerErr := semver.NewVersion(tagName)
		if innerErr != nil {
			log.Infof("skipping tag %q as it does not conform to semver %v", tagName, innerErr)
			continue
		}

		versions = append(versions, semverTag{
			tag: tag, version: v,
		})
	}

	sort.Slice(versions, func(i, j int) bool {
		// Inverted less to sort from newer to older.
		return versions[i].version.GreaterThan(versions[j].version)
	})

	if len(versions) == 0 {
		return "", nil
	}

	return versions[0].tag.Hash, nil
}
