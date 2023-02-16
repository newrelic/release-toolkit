package mapper

import (
	"fmt"
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
	checkLink func(link string) (bool, error)
}

// NewWithLeadingVCheck returns a LeadingVCheck with the provided underlying mapper and a check function which
// performs a request to the url corresponding to the link and check its status code.
func NewWithLeadingVCheck(mapper linker.Mapper) *LeadingVCheck {
	return &LeadingVCheck{mapper: mapper, checkLink: checkLinkResponse}
}

func (l *LeadingVCheck) Map(dep changelog.Dependency) string {
	link := l.mapper.Map(dep)
	if link == "" {
		log.Debug("Link check was skipped because the underlying mapper didn't return anything.")
		return link
	}

	if l.isValid((link)) {
		return link
	}

	newDep, err := l.switchDepLeadingV(dep)
	if err != nil { // it shouldn't happen unless `dep.To` is already invalid
		log.Errorf("Internal error checking links: %s", err)
		return ""
	}

	link = l.mapper.Map(newDep)
	if l.isValid((link)) {
		return link
	}

	log.Debugf("All checks failed, link for %q is omitted", dep)

	return ""
}

func (l *LeadingVCheck) switchDepLeadingV(dep changelog.Dependency) (changelog.Dependency, error) {
	literal := dep.To.Original()
	switchedLiteral := ""
	if strings.HasPrefix(literal, "v") {
		switchedLiteral = strings.TrimPrefix(literal, "v")
	} else {
		switchedLiteral = "v" + literal
	}
	switchedDep := dep
	switchedVersion, err := semver.NewVersion(switchedLiteral)
	if err != nil {
		return dep, fmt.Errorf("error switching leading v in %q: %w", literal, err)
	}

	switchedDep.To = switchedVersion

	return switchedDep, nil
}

func (l *LeadingVCheck) isValid(link string) bool {
	log.Debugf("Performing link check on %q", link)

	linkOK, err := l.checkLink(link)
	if err != nil {
		log.Errorf("The link %q could not be checked due to an unexpected error, it may be incorrect. Details: %s", link, err)
		return true
	}

	return linkOK
}

func checkLinkResponse(link string) (bool, error) {
	client := http.Client{
		Timeout: checkTimeoutSeconds * time.Second,
	}
	resp, err := client.Get(link) //nolint:noctx
	if err != nil {
		return false, fmt.Errorf("error performing the request to check %q link: %w", link, err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}
