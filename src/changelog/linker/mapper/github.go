package mapper

import (
	"fmt"
	"strings"

	"github.com/newrelic/release-toolkit/src/changelog"
	log "github.com/sirupsen/logrus"
)

const ghChangelogTemplate = "https://github.com/%s/%s/releases/tag/%s"

type Github struct{}

func (Github) Map(dep changelog.Dependency) string {
	parts := strings.Split(dep.Name, "/")

	if len(parts) < 3 || parts[0] != "github.com" {
		log.Debugf("Github mapper: Dependency %q does not match 'github.com/org/repo', is not a github dependency.", dep.Name)
		return ""
	}

	if dep.To == nil || dep.To.Original() == "" {
		log.Debugf("Github mapper: Dependency %q does not have a release version.", dep.Name)
		return ""
	}

	return fmt.Sprintf(ghChangelogTemplate, parts[1], parts[2], dep.To.Original())
}
