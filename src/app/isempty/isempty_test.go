package isempty_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"

	"github.com/newrelic/release-toolkit/src/app"
)

//nolint:paralleltest,gocyclo,cyclop
func TestIsEmpty(t *testing.T) {
	app := app.App()

	buf := &strings.Builder{}
	app.Writer = buf

	dir := t.TempDir()
	filepath := path.Join(dir, "changelog.yaml")
	file, err := os.Create(filepath)
	if err != nil {
		t.Fatalf("Error creating yaml for test: %v", err)
	}

	_, _ = file.WriteString(`
held: false 
changes: [] 
`)

	for _, tc := range []struct {
		cmd         string
		expected    string
		errExpected bool
	}{
		{
			cmd:         fmt.Sprintf("rt -yaml %s is-empty", filepath),
			expected:    "true\n",
			errExpected: false,
		},
		{
			cmd:         fmt.Sprintf("rt -gha=1 -yaml %s is-empty", filepath),
			expected:    "true\n::set-output name=is-empty::true\n",
			errExpected: false,
		},
		{
			cmd:         fmt.Sprintf("rt -gha=1 -yaml %s is-empty -fail", filepath),
			expected:    "true\n::set-output name=is-empty::true\n",
			errExpected: true,
		},
	} {
		var errValue int
		cli.OsExiter = func(code int) {
			errValue = code
		}

		buf.Reset()
		err = app.Run(strings.Fields(tc.cmd))
		if err != nil && !tc.errExpected {
			t.Fatalf("Error running app: %v", err)
		}
		if err == nil && tc.errExpected {
			t.Fatalf("An error was expected running app: %v", err)
		}

		if actual := buf.String(); actual != tc.expected {
			t.Fatalf("Expected %q, app printed: %q", tc.expected, actual)
		}

		if errValue != 1 && tc.errExpected {
			t.Fatalf("An exit code 1 was expected")
		}

		if errValue == 1 && !tc.errExpected {
			t.Fatalf("An exit code 1 was not expected")
		}
	}
}

//nolint:paralleltest
func TestIsNotEmpty(t *testing.T) {
	app := app.App()

	buf := &strings.Builder{}
	app.Writer = buf

	dir := t.TempDir()
	filepath := path.Join(dir, "changelog.yaml")
	file, err := os.Create(filepath)
	if err != nil {
		t.Fatalf("Error creating yaml for test: %v", err)
	}

	_, _ = file.WriteString(`
changes:
  - type: breaking
    message: this is broken
`)

	for _, tc := range []struct {
		cmd      string
		expected string
	}{
		{
			cmd:      fmt.Sprintf("rt -yaml %s is-empty -fail", filepath),
			expected: "false\n",
		},
		{
			cmd:      fmt.Sprintf("rt -gha=1 -yaml %s is-empty -fail", filepath),
			expected: "false\n::set-output name=is-empty::false\n",
		},
	} {
		var errValue int
		cli.OsExiter = func(code int) {
			errValue = code
		}

		buf.Reset()
		err = app.Run(strings.Fields(tc.cmd))
		if err != nil {
			t.Fatalf("Error running app: %v", err)
		}

		if actual := buf.String(); actual != tc.expected {
			t.Fatalf("Expected %q, app printed: %q", tc.expected, actual)
		}

		if errValue != 0 {
			t.Fatalf("Exit code 0 was expected")
		}
	}
}
