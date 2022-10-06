package mapper_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/changelog/linker/mapper"
)

func TestDictionary_NewDictionary(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		input    string
		expected mapper.Dictionary
	}{
		{
			input: strings.TrimSpace(`
dictionary:
  newrelic-infrastructure: "https://github.com/newrelic/nri-kubernetes/releases/tag/newrelic-infrastructure-{{.To.Original}}"
  infrastructure-agent: "https://github.com/newrelic/infrastructure-agent/releases/tag/{{.To.Original}}"
	`),
			expected: mapper.Dictionary{
				Changelogs: map[string]string{
					"newrelic-infrastructure": "https://github.com/newrelic/nri-kubernetes/releases/tag/newrelic-infrastructure-{{.To.Original}}",
					"infrastructure-agent":    "https://github.com/newrelic/infrastructure-agent/releases/tag/{{.To.Original}}",
				},
			},
		},
	} {
		actual, err := mapper.NewDictionary(strings.NewReader(tc.input))
		if err != nil {
			t.Fatalf("Error creating dictionary: %v", err)
		}
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Fatalf("Expected %v got %v", tc.expected, actual)
		}
	}
}

func TestDictionary_Map(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		input      string
		dependency changelog.Dependency
		expected   string
	}{
		{
			name: "Dependency_With_Empty_Required_Field",
			input: strings.TrimSpace(`
dictionary:
  infrastructure-agent: "https://github.com/newrelic/infrastructure-agent/releases/tag/{{.From}}"
	`),
			dependency: changelog.Dependency{
				Name: "infrastructure-agent",
				To:   semver.MustParse("1.2.3"),
			},
			expected: "",
		},
		{
			name: "Dependency_Link_Correctly_Templated",
			input: strings.TrimSpace(`
dictionary:
  infrastructure-agent: "https://github.com/newrelic/infrastructure-agent/releases/tag/{{.To.Original}}"
	`),
			dependency: changelog.Dependency{
				Name: "infrastructure-agent",
				To:   semver.MustParse("1.2.3"),
			},
			expected: "https://github.com/newrelic/infrastructure-agent/releases/tag/1.2.3",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dic, err := mapper.NewDictionary(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("Error creating dictionary: %v", err)
			}

			actual := dic.Map(tc.dependency)

			if actual != tc.expected {
				t.Fatalf("Expected %v got %v", tc.expected, actual)
			}
		})
	}
}
