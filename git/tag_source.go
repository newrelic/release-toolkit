package git

import (
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
	tags, err := s.tagsGetter.Get()
	if err != nil {
		return nil, err
	}

	return tags.Versions, nil
}
