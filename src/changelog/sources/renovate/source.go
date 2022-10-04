package renovate

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/git"
	log "github.com/sirupsen/logrus"
)

const renovateAuthor = "renovate"

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

	var dependencies []changelog.Dependency

	for _, c := range gitCommits {
		var commitDependencies []changelog.Dependency

		commitLine := strings.Split(c.Message, "\n")[0]
		// Contrary to dependabot, where we can reliably tell if a message comes from it by just inspecting the message,
		// for renovate we have to be so permissive we need to check the author to prevent false positives.
		if !strings.Contains(strings.ToLower(c.Author), renovateAuthor) {
			log.Debugf("skipping commit as it is not authored by renovate\n> %q", commitLine)
			continue
		}

		bodyDependencies := r.bodyDependencies(c.Message)
		// We later reverse the whole dependency list, but want the body dependencies to stay in the same order.
		// For this reason we reverse bodyDependencies here.
		sort.SliceStable(bodyDependencies, func(i, j int) bool {
			return j < i
		})

		commitDependencies = append(commitDependencies, bodyDependencies...)

		if len(bodyDependencies) == 0 {
			// If we do not find the dependency table in the body, we attempt to parse the title.
			commitDependencies = append(commitDependencies, r.titleDependencies(commitLine)...)
		}

		// Add commit hash and copy to dependencies
		for _, dep := range commitDependencies {
			dep.Meta.Commit = c.Hash
			dependencies = append(dependencies, dep)
		}
	}

	// Reverse order in which dependencies appear in changelog, to put the oldest first.
	// Commits are iterated in a newest-first order.
	sort.SliceStable(dependencies, func(i, j int) bool {
		return j < i
	})

	return &changelog.Changelog{Dependencies: dependencies}, nil
}

var (
	renovateRegex = regexp.MustCompile(`update (.+)`)
	prRegex       = regexp.MustCompile(`(.+) \([#!](\d+)\)$`)
	versionRegex  = regexp.MustCompile(`(.+) to (v?\d\S*)`)
)

func (r Source) titleDependencies(commitLine string) []changelog.Dependency {
	// Renovate is very flexible and its commit message very customizable. For this, we must be quite lenient with
	// the parser. Achieving that with only one regex is very hard, so we use multiple steps instead.

	// First, a very wide regex is applied:
	matches := renovateRegex.FindStringSubmatch(commitLine)
	if len(matches) == 0 {
		log.Debugf("skipping commit as it does not match renovate pattern\n> %q", commitLine)
		return nil
	}

	updateMessage := matches[1]

	// Then, we try to find the PR at the end
	var pr string
	if prMatches := prRegex.FindStringSubmatch(updateMessage); len(prMatches) != 0 {
		pr = prMatches[2]
		updateMessage = prMatches[1]
	} else {
		log.Debugf("Renovate could not extract PR number from commit %q", commitLine)
	}

	// Then, we try to see if this includes a version. If it does, we get it and we clean up everything after it.
	var updateTo *semver.Version
	if versionMatches := versionRegex.FindStringSubmatch(updateMessage); len(versionMatches) != 0 {
		updateMessage = versionMatches[1]
		updateTo, _ = semver.NewVersion(versionMatches[2])
	} else {
		log.Warnf("Renovate could not extract updated version from %q", commitLine)
	}

	// Finally, we take whatever is left from the update message and strip known prefixes and suffixes
	name := dependencyName(updateMessage)

	return []changelog.Dependency{{
		Name: name,
		To:   updateTo,
		Meta: changelog.EntryMeta{
			PR: pr,
		},
	}}
}

var (
	rowNameRegex   = regexp.MustCompile(`\[(\S+)\]`)
	rowFromToRegex = regexp.MustCompile("`(\\d\\S+)` -> `(\\d\\S+)`")
)

func (r Source) bodyDependencies(commitBody string) []changelog.Dependency {
	commitLines := strings.Split(commitBody, "\n")

	if len(commitLines) == 1 {
		log.Tracef("Skipping single-line commit as it does not have a body")
		return nil
	}

	var pr string
	if prMatches := prRegex.FindStringSubmatch(commitLines[0]); len(prMatches) != 0 {
		pr = prMatches[2]
	}

	//nolint:prealloc // Most commits won't have a commit body with dependencies.
	var bodyDeps []changelog.Dependency

	// Number of cells a row containing a dependency update has. It looks like the following example:
	// | [newrelic-infra-operator](https://hub.docker.com/r/newrelic/newrelic-infra-operator) ([source](https://togithub.com/newrelic/newrelic-infra-operator)) | patch | `1.0.7` -> `1.0.8` |
	const dependencyRowCells = 3

	for _, line := range commitLines {
		cells := strings.Split(strings.Trim(line, "| "), "|")
		if len(cells) != dependencyRowCells {
			log.Tracef("Commit line is not a renovate markdown table, skipping")
			continue
		}

		nameMatches := rowNameRegex.FindStringSubmatch(cells[0])
		if len(nameMatches) == 0 {
			log.Tracef("Dependency name not found in first table cell %q, skipping", line)
			continue
		}

		var from, to *semver.Version
		if fromTo := rowFromToRegex.FindStringSubmatch(cells[2]); len(fromTo) != 0 {
			from, _ = semver.NewVersion(fromTo[1])
			to, _ = semver.NewVersion(fromTo[2])
		} else {
			log.Warnf("Renovate could not find `from` -> `to` in %q, skipping", line)
		}

		bodyDeps = append(bodyDeps, changelog.Dependency{
			Name: nameMatches[1],
			From: from,
			To:   to,
			Meta: changelog.EntryMeta{
				PR: pr,
			},
		})
	}

	return bodyDeps
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
