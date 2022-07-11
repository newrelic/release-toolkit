package renovate

import (
	"regexp"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/changelog"
	"github.com/newrelic/release-toolkit/commit"
	"github.com/newrelic/release-toolkit/tag"
	log "github.com/sirupsen/logrus"
)

var helmReleaseRegex = regexp.MustCompile("chore\\(deps\\): update helm release (.*) to (v?(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?)")

type Extractor struct {
	tagSource    tag.Source
	commitSource commit.Source
}

func NewExtractor(tagSource tag.Source, commitSource commit.Source) Extractor {
	return Extractor{
		tagSource:    tagSource,
		commitSource: commitSource,
	}
}

func (r Extractor) Extract() ([]changelog.Dependency, error) {
	gitTags, err := r.tagSource.Tags()
	if err != nil {
		return nil, err
	}

	gitCommits, err := r.commitSource.Commits(gitTags[0].Hash)
	if err != nil {
		return nil, err
	}

	var dependencies []changelog.Dependency

	for _, c := range gitCommits {
		capturingGroups := helmReleaseRegex.FindStringSubmatch(c.Message)
		if capturingGroups == nil {
			log.Debugf("skipping commit  %s as it does not match renovate pattern", c.Message)
			continue
		}
		dependencyName := capturingGroups[1]
		dependencyTo, err := semver.NewVersion(capturingGroups[2])
		if err != nil {
			log.Debugf("skipping dependency %q as it does not conform to semver %v", dependencyName, dependencyTo)
			continue
		}

		dependencies = append(dependencies, changelog.Dependency{
			Name: dependencyName,
			To:   dependencyTo,
		})
	}
	return dependencies, nil
}