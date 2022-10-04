// Package linker implements linking dependency changelogs when they get updated.
package linker

import (
	"github.com/newrelic/release-toolkit/src/changelog"
	log "github.com/sirupsen/logrus"
)

// Linker is an object containing a map of Mappers that will be applied to Map a changelog to a dependency.
type Linker struct {
	Mappers []Mapper
}

// Mapper is any object that can link to a changelog for a dependency.
type Mapper interface {
	Map(dep changelog.Dependency) string
}

func New(mappers ...Mapper) Linker {
	return Linker{
		Mappers: mappers,
	}
}

// Link will try to Link all the dependencies in a changelog.yml to the changelogs found in the mappers.
func (l Linker) Link(cl *changelog.Changelog) error {
	for i := range cl.Dependencies {
		dep := &cl.Dependencies[i]

		if dep.Changelog != "" {
			log.Debugf("Skipping changelog linking for %q as it is already linked to %q", dep.Name, dep.Changelog)
			continue
		}

		if dep.To == nil {
			log.Debugf("Skipping changelog linking for %q as new version is unknown", dep.Name)
			continue
		}

		link := l.Map(*dep)
		if link != "" {
			dep.Changelog = link
		}
	}

	return nil
}

// Map will iterate all the mappers to map dependencies with their changelogs.
func (l Linker) Map(dep changelog.Dependency) string {
	for _, mapper := range l.Mappers {
		link := mapper.Map(dep)
		if link != "" {
			return link
		}
	}

	return ""
}
