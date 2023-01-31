package mapper

import (
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/changelog/linker"
	log "github.com/sirupsen/logrus"
)

const (
	checkTimeoutSeconds = 1
)

// LeadingVCheck wraps another mapper and performs a check over the result of Map. If the check is not
// successful, it tries to generate it prepending a leading v to the version.
type LeadingVCheck struct {
	mapper    linker.Mapper
	checkLink func(link string) bool
}

// WithLeadingVCheck returns a LeadingVCheck with the provided underlying mapper and a check function which
// performs a request to the url corresponding to the link and check its status code.
func WithLeadingVCheck(mapper linker.Mapper) *LeadingVCheck {
	return &LeadingVCheck{mapper: mapper, checkLink: checkLinkResponse}
}

func (l *LeadingVCheck) Map(dep changelog.Dependency) string {
	link := l.mapper.Map(dep)
	if link == "" || l.checkLink(link) {
		return link
	}

	newDep, changed := l.switchDepLeadingV(dep)
	if !changed {
		return ""
	}

	link = l.mapper.Map(newDep)
	if l.checkLink(link) {
		return link
	}

	return ""
}

func (l *LeadingVCheck) switchDepLeadingV(dep changelog.Dependency) (changelog.Dependency, bool) {
	literal := dep.To.Original()
	switchedLiteral := ""
	if strings.HasPrefix(literal, "v") {
		switchedLiteral = literal[1:]
	} else {
		switchedLiteral = "v" + literal
	}
	switchedDep := dep
	switchedVersion, err := semver.NewVersion(switchedLiteral)
	if err != nil {
		log.Debugf("Could not switch leading v to %q valid semver version: %s.", literal, err)
		return dep, false
	}

	switchedDep.To = switchedVersion

	return switchedDep, true
}

func checkLinkResponse(link string) bool {
	client := http.Client{
		Timeout: checkTimeoutSeconds * time.Second,
	}
	resp, err := client.Get(link) //nolint:noctx
	if err != nil {
		log.Warnf("Link %q could not be checked: %s.", link, err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
