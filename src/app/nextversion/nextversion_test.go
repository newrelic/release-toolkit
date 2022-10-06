package nextversion_test

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/newrelic/release-toolkit/src/app"
)

//nolint:paralleltest // urfave/cli cannot be tested concurrently.
func TestNextVersion_Without_Repo(t *testing.T) {
	for _, tc := range []struct {
		name       string
		yaml       string
		expected   string
		args       string
		globalargs string
	}{
		{
			name:     "Overrides_Next",
			args:     "-next v0.0.1",
			expected: "v0.0.1",
			yaml: strings.TrimSpace(`
notes: |-
    ### Important announcement (note)

    This is a release note
changes:
- type: breaking
  message: Support has been removed
- type: enhancement
  message: New feature has been added
- type: security
  message: Fixed a security issue that leaked all data
dependencies:
- name: foobar
  from: 0.0.1
  to: 0.1.0
			`),
		},
		{
			name:     "Bumps_Patch",
			args:     "-current v1.2.3",
			expected: "v1.2.4",
			yaml: strings.TrimSpace(`
changes:
- type: bugfix
  message: Just a quick fix
			`),
		},
		{
			name:     "Bumps_Minor",
			args:     "-current v1.2.3",
			expected: "v1.3.0",
			yaml: strings.TrimSpace(`
changes:
- type: enhancement
  message: New feature has been added
			`),
		},
		{
			name:     "Bumps_Major",
			args:     "-current v1.2.3",
			expected: "v2.0.0",
			yaml: strings.TrimSpace(`
changes:
- type: breaking
  message: Support has been removed
			`),
		},
		{
			name:       "Bumps_Major_GHA",
			globalargs: "-gha=true",
			args:       "-current v1.2.3",
			expected:   "v2.0.0\n::set-output name=next-version::v2.0.0",
			yaml: strings.TrimSpace(`
changes:
- type: breaking
  message: Support has been removed
			`),
		},
	} {
		tc := tc
		//nolint:paralleltest // urfave/cli cannot be tested concurrently.
		t.Run(tc.name, func(t *testing.T) {
			tDir := t.TempDir()

			app := app.App()

			yamlPath := path.Join(tDir, "changelog.yaml")
			yamlFile, err := os.Create(yamlPath)
			if err != nil {
				t.Fatalf("Error creating yaml for test: %v", err)
			}
			_, _ = yamlFile.WriteString(tc.yaml)
			_ = yamlFile.Close()

			buf := &strings.Builder{}
			app.Writer = buf

			err = app.Run(strings.Fields(fmt.Sprintf("rt -yaml %s %s next-version %s", yamlPath, tc.globalargs, tc.args)))
			if err != nil {
				t.Fatalf("Error running app: %v", err)
			}

			if actual := buf.String(); actual != tc.expected+"\n" {
				t.Fatalf("Expected %q, got %q", tc.expected, actual)
			}
		})
	}
}

//nolint:funlen,paralleltest // urfave/cli cannot be tested concurrently.
func TestNextVersion(t *testing.T) {
	tags := []string{
		"v0.1.0",
		"v1.0.0",
		"v2.0.0", // Unordered on purpose, this should be the current version.
		"v1.1.0",
		"v1.2.0",
		"chart-5.0.0", // This should be the current version if prefix is 'chart-'
		"chart-4.0.0",
		"chart-4.1.0-beta",
	}

	for _, tc := range []struct {
		name     string
		yaml     string
		args     string
		expected string
	}{
		{
			name:     "Overrides_Next",
			args:     "--next v0.0.1",
			expected: "v0.0.1",
			yaml: strings.TrimSpace(`
notes: |-
    ### Important announcement (note)

    This is a release note
changes:
- type: breaking
  message: Support has been removed
- type: enhancement
  message: New feature has been added
- type: security
  message: Fixed a security issue that leaked all data
dependencies:
- name: foobar
  from: 0.0.1
  to: 0.1.0
			`),
		},
		{
			name:     "Overrides_Current",
			expected: "v3.0.0",
			yaml: strings.TrimSpace(`
notes: |-
    ### Important announcement (note)

    This is a release note
changes:
- type: breaking
  message: Support has been removed
- type: enhancement
  message: New feature has been added
- type: security
  message: Fixed a security issue that leaked all data
dependencies:
- name: foobar
  from: 0.0.1
  to: 0.1.0
			`),
		},
		{
			name:     "Bumps_Major",
			expected: "v3.0.0",
			yaml: strings.TrimSpace(`
notes: |-
    ### Important announcement (note)

    This is a release note
changes:
- type: breaking
  message: Support has been removed
- type: enhancement
  message: New feature has been added
- type: security
  message: Fixed a security issue that leaked all data
dependencies:
- name: foobar
  from: 0.0.1
  to: 0.1.0
			`),
		},
		{
			name:     "Bumps_Patch",
			expected: "v2.0.1",
			yaml: strings.TrimSpace(`
changes:
- type: bugfix
  message: Just a quick fix
			`),
		},
		{
			name:     "Bumps_Minor",
			expected: "v2.1.0",
			yaml: strings.TrimSpace(`
changes:
- type: enhancement
  message: New feature has been added
			`),
		},
		{
			name:     "Bumps_Chart_Minor",
			expected: "v5.1.0",
			args:     "--tag-prefix chart-",
			yaml: strings.TrimSpace(`
changes:
- type: enhancement
  message: New feature has been added
			`),
		},
	} {
		tc := tc
		//nolint:paralleltest // urfave/cli cannot be tested concurrently.
		t.Run(tc.name, func(t *testing.T) {
			repoDir := repoWithTags(t, tags...)

			app := app.App()

			yamlPath := path.Join(repoDir, "changelog.yaml")
			yamlFile, err := os.Create(yamlPath)
			if err != nil {
				t.Fatalf("Error creating yaml for test: %v", err)
			}
			_, _ = yamlFile.WriteString(tc.yaml)
			_ = yamlFile.Close()

			buf := &strings.Builder{}
			app.Writer = buf

			err = app.Run(strings.Fields(fmt.Sprintf("rt -yaml %s next-version -git-root %s %s", yamlPath, repoDir, tc.args)))
			if err != nil {
				t.Fatalf("Error running app: %v", err)
			}

			if actual := buf.String(); actual != tc.expected+"\n" {
				t.Fatalf("Expected %q, got %q", tc.expected, actual)
			}
		})
	}
}

func repoWithTags(t *testing.T, tags ...string) string {
	t.Helper()

	dir := t.TempDir()

	cmds := []string{
		"git init",
		"git config user.email test@user.tld",
		"git config user.name Test",
		"git config commit.gpgsign false",
		"touch a",
		"git add a",
		"git commit -m test",
	}

	for _, t := range tags {
		cmds = append(cmds, fmt.Sprintf("git tag %s", t))
	}

	for _, cmdline := range cmds {
		cmdparts := strings.Fields(cmdline)
		//nolint:gosec // This is a test, we trust hardcoded input.
		cmd := exec.Command(cmdparts[0], cmdparts[1:]...)
		cmd.Dir = dir

		out := strings.Builder{}
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			t.Errorf("%s output:\n%s", cmdline, out.String())
			t.Fatalf("Error bootstraping test git repo: %v", err)
		}
	}

	return dir
}
