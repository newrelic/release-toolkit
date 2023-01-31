package mapper

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/changelog/linker"
	"github.com/stretchr/testify/assert"
)

type mapperMock struct {
	link string
}

func (m *mapperMock) Map(dep changelog.Dependency) string {
	if m.link == "" {
		return ""
	}
	return m.link + "-" + dep.To.Original()
}

func TestLeadingVCheck_Map(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		wrapped     linker.Mapper
		dep         changelog.Dependency
		checkValue  string
		expected    string
		shouldCheck bool
	}{
		{
			name:        "Link is empty",
			wrapped:     &mapperMock{link: ""},
			dep:         changelog.Dependency{To: semver.MustParse("1.2.3")},
			expected:    "",
			shouldCheck: false,
		},
		{
			name:        "Check passed with leading v",
			wrapped:     &mapperMock{link: "link"},
			dep:         changelog.Dependency{To: semver.MustParse("v1.2.3")},
			checkValue:  "link-v1.2.3",
			expected:    "link-v1.2.3",
			shouldCheck: true,
		},
		{
			name:        "Check passed without leading v",
			wrapped:     &mapperMock{link: "link"},
			dep:         changelog.Dependency{To: semver.MustParse("1.2.3")},
			checkValue:  "link-1.2.3",
			expected:    "link-1.2.3",
			shouldCheck: true,
		},
		{
			name:        "Needs prepending v to pass the check",
			wrapped:     &mapperMock{link: "link"},
			dep:         changelog.Dependency{To: semver.MustParse("1.2.3")},
			checkValue:  "link-v1.2.3",
			expected:    "link-v1.2.3",
			shouldCheck: true,
		},
		{
			name:        "Needs removing v to pass the check",
			wrapped:     &mapperMock{link: "link"},
			dep:         changelog.Dependency{To: semver.MustParse("v1.2.3")},
			checkValue:  "link-1.2.3",
			expected:    "link-1.2.3",
			shouldCheck: true,
		},
		{
			name:        "Check will not pass even prepending v",
			wrapped:     &mapperMock{link: "link"},
			dep:         changelog.Dependency{To: semver.MustParse("1.2.3")},
			checkValue:  "no-way",
			expected:    "",
			shouldCheck: true,
		},
		{
			name:        "Check will not pass even removing v",
			wrapped:     &mapperMock{link: "link"},
			dep:         changelog.Dependency{To: semver.MustParse("v1.2.3")},
			checkValue:  "no-way",
			expected:    "",
			shouldCheck: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			checkCalled := false
			checkLink := func(link string) bool {
				checkCalled = true
				return tc.checkValue == link
			}

			m := &LeadingVCheck{mapper: tc.wrapped, checkLink: checkLink}

			assert.Equal(t, tc.expected, m.Map(tc.dep))
			assert.Equal(t, tc.shouldCheck, checkCalled)
		})
	}
}

func TestLeadingVCheck_switchDepLeadingV(t *testing.T) {
	t.Parallel()

	m := &LeadingVCheck{}

	testCases := []struct {
		Name     string
		Original changelog.Dependency
		changed  bool
		Expected changelog.Dependency
	}{
		{
			Name:     "With leading v",
			Original: changelog.Dependency{To: semver.MustParse("v1.2.3")},
			Expected: changelog.Dependency{To: semver.MustParse("1.2.3")},
			changed:  true,
		},
		{
			Name:     "Can include leading v",
			Original: changelog.Dependency{To: semver.MustParse("1.2.3")},
			Expected: changelog.Dependency{To: semver.MustParse("v1.2.3")},
			changed:  true,
		},
		{
			Name:     "Error including leading v",
			Original: changelog.Dependency{To: &semver.Version{}},
			changed:  false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			got, couldCreate := m.switchDepLeadingV(tc.Original)
			assert.Equal(t, tc.changed, couldCreate)
			if couldCreate {
				assert.Equal(t, tc.Expected, got)
			}
		})
	}
}

func TestLeadingVCheck_checkLink(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name    string
		handler http.HandlerFunc
		OK      bool
	}{
		{
			Name: "Request OK",
			OK:   true,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			Name: "Not found",
			OK:   false,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
		},
		{
			Name: "Timeout",
			OK:   false,
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Second)
				w.WriteHeader(http.StatusOK)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			assert.Equal(t, tc.OK, checkLinkResponse(server.URL))
		})
	}
}
