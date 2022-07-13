package git

import (
	"fmt"

	"github.com/Masterminds/semver"
)

// TagsSource implements the `tag.Source` interface, using tags from a git repository as a source for previous versions.
type TagsSource struct {
	tagsGetter SemverTagsGetter
}

func NewTagsSource(tagsGetter SemverTagsGetter) *TagsSource {
	return &TagsSource{tagsGetter: tagsGetter}
}

func (s *TagsSource) Versions() ([]*semver.Version, error) {
	tags, err := s.tagsGetter.Tags()
	if err != nil {
		return nil, fmt.Errorf("getting tags: %w", err)
	}

	return tags.Versions, nil
}
