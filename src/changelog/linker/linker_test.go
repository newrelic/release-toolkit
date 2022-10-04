package linker_test

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/changelog/linker"
	"github.com/newrelic/release-toolkit/src/changelog/linker/mapper"
)

func TestLinker_Link(t *testing.T) {
	t.Parallel()

	dictionary := map[string]string{
		"infrastructure-agent": "https://github.com/newrelic/infrastructure-agent/releases/tag/{{.To.Original}}",
	}

	link := linker.New(mapper.Dictionary{Changelogs: dictionary}, mapper.Github{})

	for _, tc := range []struct {
		name      string
		changelog *changelog.Changelog
		expected  map[string]string
	}{
		{
			name: "Already_Present_Dependency_Changes_Do_Not_Get_Linked",
			changelog: &changelog.Changelog{
				Changes: []changelog.Entry{
					{Message: "Change one", Type: changelog.TypeBugfix},
				},
				Dependencies: []changelog.Dependency{
					{Name: "infrastructure-agent", To: semver.MustParse("v1.2.3"), Changelog: "https://www.newrelic.com"},
					{Name: "github.com/spf13/viper", To: semver.MustParse("v2.2.3"), Changelog: "https://www.newrelic.com"},
				},
			},
			expected: map[string]string{
				"infrastructure-agent":   "https://www.newrelic.com",
				"github.com/spf13/viper": "https://www.newrelic.com",
			},
		},
		{
			name: "Dependency_Changes_Get_Linked",
			changelog: &changelog.Changelog{
				Changes: []changelog.Entry{
					{Message: "Change one", Type: changelog.TypeBugfix},
				},
				Dependencies: []changelog.Dependency{
					{Name: "infrastructure-agent", To: semver.MustParse("v1.2.3")},
					{Name: "github.com/spf13/viper", To: semver.MustParse("v2.2.3")},
				},
			},
			expected: map[string]string{
				"infrastructure-agent":   "https://github.com/newrelic/infrastructure-agent/releases/tag/v1.2.3",
				"github.com/spf13/viper": "https://github.com/spf13/viper/releases/tag/v2.2.3",
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := link.Link(tc.changelog)
			if err != nil {
				t.Fatalf("Err linking: %v", err)
			}

			for _, dep := range tc.changelog.Dependencies {
				if dep.Changelog != tc.expected[dep.Name] {
					t.Fatalf("Dependency %s changelog not matching: %s != %s", dep.Name, dep.Changelog, tc.expected[dep.Name])
				}
			}
		})
	}
}
