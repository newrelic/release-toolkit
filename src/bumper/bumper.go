package bumper

import (
	"errors"
	"fmt"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/src/bump"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/version"
)

var (
	ErrEmptySource = errors.New("could not find any existing version")
	ErrNoNewVersion = errors.New("no newer version to be bumped")
)


// Bumper takes a changelog and a version and figures out the next one.
type Bumper struct {
	changelog     changelog.Changelog
	EntryCap      bump.Type
	DependencyCap bump.Type
}

// New creates a new bumper.
func New(c changelog.Changelog) Bumper {
	return Bumper{
		changelog:     c,
		EntryCap:      bump.Major,
		DependencyCap: bump.Major,
	}
}

// Bump uses the Bumper's changelog.Changelog to compute the next version from v.
func (b Bumper) Bump(v *semver.Version) *semver.Version {
	entryBump := bump.None
	for _, e := range b.changelog.Changes {
		entryBump = entryBump.With(e.BumpType())
	}
	entryBump = entryBump.Cap(b.EntryCap)

	dependencyBump := bump.None
	for _, d := range b.changelog.Dependencies {
		dependencyBump = dependencyBump.With(d.BumpType())
	}
	dependencyBump = dependencyBump.Cap(b.DependencyCap)

	return bump.Bump(v, entryBump.With(dependencyBump))
}

// BumpSource operates just like Bump, except it extracts tags from the supplied tag.Source and applies the bump
// on the latest (in semver order) version it finds.
func (b Bumper) BumpSource(source version.Source) (*semver.Version, error) {
	versions, err := source.Versions()
	if err != nil {
		return nil, fmt.Errorf("getting versions from source: %w", err)
	}

	if len(versions) == 0 {
		return nil, ErrEmptySource
	}

	sort.Slice(versions, func(i, j int) bool {
		// We use GreaterThan as lessFunc to sort from largest to smallest.
		return versions[i].GreaterThan(versions[j])
	})

	nextVersion := b.Bump(versions[0])
	if versions[0] == nextVersion {
		return nil, ErrNoNewVersion
	}

	return nextVersion, nil
}
