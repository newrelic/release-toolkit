package renovate

import (
	"fmt"
	"regexp"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/changelog"
	"github.com/newrelic/release-toolkit/git"
	log "github.com/sirupsen/logrus"
)

var commitRegex = regexp.MustCompile(`update .* ([\w-/.]+) to ([^\s*]+)(?: \(([^\s]+)\))?`)

type Source struct {
	tagsReleaseGetter git.TagsReleaseGetter
	commitsGetter     git.CommitsGetter
}

func NewSource(tagsReleaseGetter git.TagsReleaseGetter, commitsGetter git.CommitsGetter) Source {
	return Source{
		tagsReleaseGetter: tagsReleaseGetter,
		commitsGetter:     commitsGetter,
	}
}

func (r Source) Source() ([]changelog.Dependency, error) {
	lastHash, err := r.tagsReleaseGetter.LastReleaseHash()
	if err != nil {
		return nil, fmt.Errorf("getting last release hash: %w", err)
	}

	gitCommits, err := r.commitsGetter.Commits(lastHash)
	if err != nil {
		return nil, fmt.Errorf("getting commits: %w", err)
	}

	dependencies := make([]changelog.Dependency, 0)

	for _, c := range gitCommits {
		capturingGroups := commitRegex.FindStringSubmatch(c.Message)
		if capturingGroups == nil {
			log.Debugf("skipping commit  %s as it does not match renovate pattern", c.Message)
			continue
		}
		dependencyName := capturingGroups[1]
		dependencyTo, err := semver.NewVersion(capturingGroups[2])
		if err != nil {
			log.Debugf("skipping dependency %q as it doesn't conform to semver %v", dependencyName, dependencyTo)
			continue
		}

		dependencies = append(dependencies, changelog.Dependency{
			Name: dependencyName,
			To:   dependencyTo,
			Meta: changelog.EntryMeta{
				Author: c.Author,
				PR:     capturingGroups[3],
				Commit: c.Hash,
			},
		})
	}
	return dependencies, nil
}
