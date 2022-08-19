package renovate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/changelog"
	"github.com/newrelic/release-toolkit/git"
	log "github.com/sirupsen/logrus"
)

const renovateAuthor = "renovate"

var (
	renovateRegex = regexp.MustCompile(`update (.+)`)
	prRegex       = regexp.MustCompile(`(.+) \([#!](\d+)\)$`)
	versionRegex  = regexp.MustCompile(`(.+) to (v?\d\S*)`)
)

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

	log.Debugf("Listing commits until last tag %q", lastHash)
	gitCommits, err := r.commitsGetter.Commits(lastHash)
	if err != nil {
		return nil, fmt.Errorf("getting commits: %w", err)
	}
	if len(gitCommits) == 0 {
		log.Infof("Renovate source did not find any commit since %q", lastHash)
	}

	dependencies := make([]changelog.Dependency, 0)

	for _, c := range gitCommits {
		// Renovate is very flexible and its commit message very customizable. For this, we must be quite lenient with
		// the parser. Achieving that with only one regex is very hard, so we use multiple steps instead.

		commitLine := strings.Split(c.Message, "\n")[0]

		// Contrary to dependabot, where we can reliably tell if a message comes from it by just inspecting the message,
		// for renovate we have to be so permissive we need to check the author to prevent false positives.
		if !strings.Contains(strings.ToLower(c.Author), renovateAuthor) {
			log.Debugf("skipping commit as it is not authored by renovate\n> %q", commitLine)
			continue
		}

		// First, a very wide regex is applied:
		matches := renovateRegex.FindStringSubmatch(commitLine)
		if len(matches) == 0 {
			log.Debugf("skipping commit as it does not match renovate pattern\n> %q", commitLine)
			continue
		}

		updateMessage := matches[1]

		// Then, we try to find the PR at the end
		var pr string
		if prMatches := prRegex.FindStringSubmatch(updateMessage); len(prMatches) != 0 {
			pr = prMatches[2]
			updateMessage = prMatches[1]
		}

		// Then, we try to see if this includes a version. If it does, we get it and we clean up everything after it.
		var updateTo *semver.Version
		if versionMatches := versionRegex.FindStringSubmatch(updateMessage); len(versionMatches) != 0 {
			updateMessage = versionMatches[1]
			updateTo, _ = semver.NewVersion(versionMatches[2])
		}

		// Finally, we take whatever is left from the update message and strip known prefixes and suffixes
		name := dependencyName(updateMessage)

		dependencies = append(dependencies, changelog.Dependency{
			Name: name,
			To:   updateTo,
			Meta: changelog.EntryMeta{
				PR:     pr,
				Commit: c.Hash,
			},
		})
	}

	// Reverse order in which dependencies appear in changelog, to put the oldest first.
	// Commits are iterated in a newest-first order.
	nDeps := len(dependencies)
	sortedDependencies := make([]changelog.Dependency, nDeps)
	for i := 0; i < nDeps; i++ {
		sortedDependencies[nDeps-1-i] = dependencies[i]
	}

	return &changelog.Changelog{Dependencies: sortedDependencies}, nil
}

func dependencyName(rawName string) string {
	knownManagerAffixes := []string{
		"helm release",
		"module",
		"docker tag",
		"action",
		"dependency",
	}

	rawName = strings.ToLower(rawName)

	for _, affix := range knownManagerAffixes {
		// Replace affixes either at the beginning or the end of the string.
		// This prevents removing e.g. `module` from a dependency name such as my-module.
		rawName = strings.TrimPrefix(rawName, affix+" ")
		rawName = strings.TrimSuffix(rawName, " "+affix)
	}

	return strings.TrimSpace(rawName)
}
