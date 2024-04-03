package mapper

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/changelog/linker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

//nolint:funlen,goconst,goerr113
func TestLeadingVCheck_Map(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		wrapped     linker.Mapper
		dep         changelog.Dependency
		checkLink   func(string) (bool, error)
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
			name:    "Check passed with leading v",
			wrapped: &mapperMock{link: "link"},
			dep:     changelog.Dependency{To: semver.MustParse("v1.2.3")},
			checkLink: func(link string) (bool, error) {
				return link == "link-v1.2.3", nil
			},
			expected:    "link-v1.2.3",
			shouldCheck: true,
		},
		{
			name:    "Check passed without leading v",
			wrapped: &mapperMock{link: "link"},
			dep:     changelog.Dependency{To: semver.MustParse("1.2.3")},
			checkLink: func(link string) (bool, error) {
				return link == "link-1.2.3", nil
			},
			expected:    "link-1.2.3",
			shouldCheck: true,
		},
		{
			name:    "Needs prepending v to pass the check",
			wrapped: &mapperMock{link: "link"},
			dep:     changelog.Dependency{To: semver.MustParse("1.2.3")},
			checkLink: func(link string) (bool, error) {
				return link == "link-v1.2.3", nil
			},
			expected:    "link-v1.2.3",
			shouldCheck: true,
		},
		{
			name:    "Needs removing v to pass the check",
			wrapped: &mapperMock{link: "link"},
			dep:     changelog.Dependency{To: semver.MustParse("v1.2.3")},
			checkLink: func(link string) (bool, error) {
				return link == "link-1.2.3", nil
			},
			expected:    "link-1.2.3",
			shouldCheck: true,
		},
		{
			name:    "Check will not pass even prepending v",
			wrapped: &mapperMock{link: "link"},
			dep:     changelog.Dependency{To: semver.MustParse("1.2.3")},
			checkLink: func(_ string) (bool, error) {
				return false, nil
			},
			expected:    "",
			shouldCheck: true,
		},
		{
			name:    "Check will not pass even removing v",
			wrapped: &mapperMock{link: "link"},
			dep:     changelog.Dependency{To: semver.MustParse("v1.2.3")},
			checkLink: func(_ string) (bool, error) {
				return false, nil
			},
			expected:    "",
			shouldCheck: true,
		},
		{
			name:    "Check returns an error",
			wrapped: &mapperMock{link: "link"},
			dep:     changelog.Dependency{To: semver.MustParse("1.2.3")},
			checkLink: func(_ string) (bool, error) {
				return false, errors.New("")
			},
			expected:    "link-1.2.3", // It uses the link from the underlying mapper.
			shouldCheck: true,
		},
		{
			name:    "Check returns an error after switching",
			wrapped: &mapperMock{link: "link"},
			dep:     changelog.Dependency{To: semver.MustParse("1.2.3")},
			checkLink: func(link string) (bool, error) {
				if link == "link-1.2.3" {
					return false, nil
				}
				return false, errors.New("")
			},
			expected:    "link-v1.2.3", // It uses the link from the underlying mapper.
			shouldCheck: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			checkCalled := false
			checkLink := func(link string) (bool, error) {
				checkCalled = true
				return tc.checkLink(link)
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
		name     string
		original changelog.Dependency
		err      bool
		expected changelog.Dependency
	}{
		{
			name:     "With leading v",
			original: changelog.Dependency{To: semver.MustParse("v1.2.3")},
			expected: changelog.Dependency{To: semver.MustParse("1.2.3")},
		},
		{
			name:     "Without leading v",
			original: changelog.Dependency{To: semver.MustParse("1.2.3")},
			expected: changelog.Dependency{To: semver.MustParse("v1.2.3")},
		},
		{
			name:     "Error including leading v",
			original: changelog.Dependency{To: &semver.Version{}},
			err:      true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := m.switchDepLeadingV(tc.original)
			if tc.err {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestLeadingVCheck_checkLink(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		handler http.HandlerFunc
		ok      bool
		err     bool
	}{
		{
			name: "Request OK",
			ok:   true,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "Not found",
			ok:   false,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
		},
		{
			name: "Timeout",
			ok:   false,
			err:  true,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				time.Sleep(2 * time.Second)
				w.WriteHeader(http.StatusOK)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			result, err := checkLinkResponse(server.URL)
			if tc.err {
				require.Error(t, err)
				return
			}
			assert.Equal(t, tc.ok, result)
		})
	}
}
