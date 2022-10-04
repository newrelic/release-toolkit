package mapper

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/src/changelog"
)

func TestGithub_Map(t *testing.T) {
	t.Parallel()

	github := Github{}

	for _, tc := range []struct {
		name       string
		dependency changelog.Dependency
		expected   string
	}{
		{
			name: "Not_Github_Dependency_Does_Not_Get_Templated",
			dependency: changelog.Dependency{
				Name: "gitlab.com/spf13/viper",
				To:   semver.MustParse("v1.12.0"),
			},
			expected: "",
		},
		{
			name: "Wrong_Format_Github_Dependency_Does_Not_Get_Templated",
			dependency: changelog.Dependency{
				Name: "github.com/spf13",
				To:   semver.MustParse("v1.12.0"),
			},
			expected: "",
		},
		{
			name: "Github_Dependency_Without_Version_Does_Not_Get_Templated",
			dependency: changelog.Dependency{
				Name: "github.com/spf13/viper",
			},
			expected: "",
		},
		{
			name: "Github dependency gets templated",
			dependency: changelog.Dependency{
				Name: "github.com/spf13/viper",
				To:   semver.MustParse("v1.12.0"),
			},
			expected: "https://github.com/spf13/viper/releases/tag/v1.12.0",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			actual := github.Map(tc.dependency)
			if actual != tc.expected {
				t.Fatalf("Link map not matching: %s != %s", actual, tc.expected)
			}
		})
	}
}
