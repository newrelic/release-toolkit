package link_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/newrelic/release-toolkit/src/app"
)

//nolint:paralleltest,funlen // urfave/cli cannot be tested concurrently.
func TestLink(t *testing.T) {
	for _, tc := range []struct {
		name       string
		chlog      string
		dictionary string
		expected   string
	}{
		{
			name: "Dictionary_Takes_Precedence",
			chlog: strings.TrimSpace(`
notes: ""
changes: []
dependencies:
- name: github.com/spf13/viper
  from: 4.0.3
  to: 4.1.2
			`),
			dictionary: strings.TrimSpace(`
dictionary:
  github.com/spf13/viper: "https://github.com/spf13/viper/releases/tag/my-custom-viper-{{.To.Original}}"
			`),
			expected: strings.TrimLeft(`
notes: ""
changes: []
dependencies:
    - name: github.com/spf13/viper
      from: 4.0.3
      to: 4.1.2
      changelog: https://github.com/spf13/viper/releases/tag/my-custom-viper-4.1.2
`, "\n"),
		},
		{
			name: "Adds_Dependency_Changelogs",
			chlog: strings.TrimSpace(`
notes: |-
    ### Important announcement (note)

    This is a release note
changes:
- type: breaking
  message: Support has been removed
dependencies:
- name: newrelic-infrastructure
  from: 1.0.1
  to: 2.1.0
- name: infrastructure-agent
  from: 0.0.1
  to: 0.1.0
- name: github.com/spf13/viper
  from: 4.0.3
  to: 4.1.2
- name: foobar
  from: 1.0.0
  to: 2.0.0
			`),
			dictionary: strings.TrimSpace(`
dictionary:
  newrelic-infrastructure: "https://github.com/newrelic/nri-kubernetes/releases/tag/newrelic-infrastructure-{{.To.Original}}"
  infrastructure-agent: "https://github.com/newrelic/infrastructure-agent/releases/tag/{{.To.Original}}"
			`),
			expected: strings.TrimLeft(`
notes: |-
    ### Important announcement (note)

    This is a release note
changes:
    - type: breaking
      message: Support has been removed
dependencies:
    - name: newrelic-infrastructure
      from: 1.0.1
      to: 2.1.0
      changelog: https://github.com/newrelic/nri-kubernetes/releases/tag/newrelic-infrastructure-2.1.0
    - name: infrastructure-agent
      from: 0.0.1
      to: 0.1.0
      changelog: https://github.com/newrelic/infrastructure-agent/releases/tag/0.1.0
    - name: github.com/spf13/viper
      from: 4.0.3
      to: 4.1.2
      changelog: https://github.com/spf13/viper/releases/tag/4.1.2
    - name: foobar
      from: 1.0.0
      to: 2.0.0
`, "\n"),
		},
	} {
		tc := tc
		//nolint:paralleltest // urfave/cli cannot be tested concurrently.
		t.Run(tc.name, func(t *testing.T) {
			tDir := t.TempDir()

			app := app.App()

			chlogPath := path.Join(tDir, "changelog.yaml")
			chlogFile, err := os.Create(chlogPath)
			if err != nil {
				t.Fatalf("Error creating yaml for test: %v", err)
			}
			_, _ = chlogFile.WriteString(tc.chlog)
			_ = chlogFile.Close()

			dicPath := path.Join(tDir, "dictionary.yaml")
			dicFile, err := os.Create(dicPath)
			if err != nil {
				t.Fatalf("Error creating yaml for test: %v", err)
			}
			_, _ = dicFile.WriteString(tc.dictionary)
			_ = dicFile.Close()

			buf := &strings.Builder{}
			app.Writer = buf

			err = app.Run(strings.Fields(fmt.Sprintf("rt -yaml %s link-changelog -dictionary %s", chlogPath, dicPath)))
			if err != nil {
				t.Fatalf("Error running app: %v", err)
			}

			actual, err := os.ReadFile(chlogPath)
			if err != nil {
				t.Fatalf("Error reading changelog file: %v", err)
			}

			if diff := cmp.Diff(tc.expected, string(actual)); diff != "" {
				t.Fatalf("Changelog.yml is not as expected\n%s", diff)
			}
		})
	}
}

func TestSample(t *testing.T) {
	t.Parallel()

	expectedSample := strings.TrimLeft(`
dictionary:
    golangci-lint: https://github.com/golangci/golangci-lint/releases/tag/{{.To.Original}}
    newrelic-infrastructure: https://github.com/newrelic/nri-kubernetes/releases/tag/newrelic-infrastructure-{{.To.Original}}
`, "\n")

	app := app.App()
	buf := &strings.Builder{}
	app.Writer = buf

	err := app.Run(strings.Fields("rt link-changelog --sample"))
	if err != nil {
		t.Fatalf("Error running app: %v", err)
	}

	if diff := cmp.Diff(expectedSample, buf.String()); diff != "" {
		t.Fatalf("Changelog.yml is not as expected\n%s", diff)
	}
}
