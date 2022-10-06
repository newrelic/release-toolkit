package dependabot

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/git"
	log "github.com/sirupsen/logrus"
)

var commitRegex = regexp.MustCompile(`(?m)[Bb]ump (\S+)(?: from (\S+))?(?: to (\S+))?(?:.+\([#!](\d+)\)$)?`)

type Source struct {
	tagsVersionGetter git.TagsVersionGetter
	commitsGetter     git.CommitsGetter
}

func NewSource(tagsVersionGetter git.TagsVersionGetter, commitsGetter git.CommitsGetter) Source {
	return Source{
		tagsVersionGetter: tagsVersionGetter,
		commitsGetter:     commitsGetter,
	}
}

func (r Source) Changelog() (*changelog.Changelog, error) {
	lastHash, err := r.tagsVersionGetter.LastVersionHash()
	if err != nil {
		return nil, fmt.Errorf("getting last version hash: %w", err)
	}

	gitCommits, err := r.commitsGetter.Commits(lastHash)
	if err != nil {
		return nil, fmt.Errorf("getting commits: %w", err)
	}
	if len(gitCommits) == 0 {
		log.Infof("Dependabot source did not find any commit since %q", lastHash)
	}

	dependencies := make([]changelog.Dependency, 0)

	for _, c := range gitCommits {
		capturingGroups := commitRegex.FindStringSubmatch(c.Message)
		if len(capturingGroups) == 0 {
			log.Debugf("skipping commit  %s as it does not match dependabot pattern", c.Message)
			continue
		}

		dependencyName := capturingGroups[1]
		dependencyFrom, err := semver.NewVersion(capturingGroups[2])
		if err != nil {
			log.Debugf("skipping dependency %q as it doesn't conform to semver %v", dependencyName, dependencyFrom)
		}
		dependencyTo, err := semver.NewVersion(capturingGroups[3])
		if err != nil {
			log.Debugf("skipping dependency %q as it doesn't conform to semver %v", dependencyName, dependencyTo)
		}

		dependencies = append(dependencies, changelog.Dependency{
			Name: dependencyName,
			From: dependencyFrom,
			To:   dependencyTo,
			Meta: changelog.EntryMeta{
				PR:     capturingGroups[4],
				Commit: c.Hash,
			},
		})
	}

	// Reverse order in which dependencies appear in changelog, to put the oldest first.
	// Commits are iterated in a newest-first order.
	sort.SliceStable(dependencies, func(i, j int) bool {
		return j < i
	})

	return &changelog.Changelog{Dependencies: dependencies}, nil
}
