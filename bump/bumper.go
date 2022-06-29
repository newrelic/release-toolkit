package bump

import (
	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/bump/bumptype"
	"github.com/newrelic/release-toolkit/changelog"
)

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

func (b Bumper) Bump(v *semver.Version) *semver.Version {
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

	return bumptype.Bump(v, entryBump.With(dependencyBump))
}
