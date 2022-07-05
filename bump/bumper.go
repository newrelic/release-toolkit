package bump

import (
	"errors"
	"fmt"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/bump/bumptype"
	"github.com/newrelic/release-toolkit/changelog"
	"github.com/newrelic/release-toolkit/tag"
)

var ErrNoTags = errors.New("source did not return any tag")

// Bumper takes a changelog and a version and figures out the next one.
type Bumper struct {
	changelog     changelog.Changelog
	EntryCap      bumptype.Type
	DependencyCap bumptype.Type
}

// NewBumper creates a new bumper.
func NewBumper(c changelog.Changelog) Bumper {
	return Bumper{
		changelog:     c,
		EntryCap:      bumptype.Major,
		DependencyCap: bumptype.Major,
	}
}

// Bump uses the Bumper's changelog.Changelog to compute the next version from v.
func (b Bumper) Bump(v *semver.Version) (*semver.Version, error) {
	entryBump := bumptype.None
	for _, e := range b.changelog.Changes {
		entryBump = entryBump.With(e.BumpType())
	}
	entryBump = entryBump.Cap(b.EntryCap)

	dependencyBump := bumptype.None
	for _, d := range b.changelog.Dependencies {
		dependencyBump = dependencyBump.With(d.BumpType())
	}
	dependencyBump = dependencyBump.Cap(b.EntryCap)

	return bumptype.Bump(v, entryBump.With(dependencyBump)), nil
}

// BumpSource operates just like Bump, except it extracts tags from the supplied tag.Source and applies the bump
// on the latest (in semver order) version it finds.
func (b Bumper) BumpSource(source tag.Source) (*semver.Version, error) {
	versions, err := source.Tags()
	if err != nil {
		return nil, fmt.Errorf("getting versions from source: %w", err)
	}

	if len(versions) == 0 {
		return nil, ErrNoTags
	}

	sort.Slice(versions, func(i, j int) bool {
		// We use GreaterThan as lessFunc to sort from largest to smallest.
		return versions[i].GreaterThan(versions[j])
	})

	return b.Bump(versions[0])
}
