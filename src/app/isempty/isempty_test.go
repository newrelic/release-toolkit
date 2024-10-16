package isempty_test

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"

	"github.com/newrelic/release-toolkit/internal/testutil"
	"github.com/newrelic/release-toolkit/src/app"
	"github.com/newrelic/release-toolkit/src/app/gha"
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
		cmd               string
		expectedStdOutput string
		expectedGHAOutput string
		errExpected       bool
	}{
		{
			cmd:               fmt.Sprintf("rt -yaml %s is-empty", filepath),
			expectedStdOutput: "true\n",
			errExpected:       false,
		},
		{
			cmd:               fmt.Sprintf("rt -gha=1 -yaml %s is-empty", filepath),
			expectedStdOutput: "true\n",
			expectedGHAOutput: "is-empty=true\n",
			errExpected:       false,
		},
		{
			cmd:               fmt.Sprintf("rt -gha=1 -yaml %s is-empty -fail", filepath),
			expectedStdOutput: "true\n",
			expectedGHAOutput: "is-empty=true\n",
			errExpected:       true,
		},
	} {
		var errValue int
		cli.OsExiter = func(code int) {
			errValue = code
		}

		ghaOutputFileName := path.Join(t.TempDir(), "temporary_github_output_file")
		t.Setenv(gha.GithubOutput, ghaOutputFileName)

		ghaOutputFile, err := os.Create(ghaOutputFileName)
		if err != nil {
			t.Fatalf("Error creating temporary GHA output file for test: %v", err)
		}

		buf.Reset()
		err = app.Run(strings.Fields(tc.cmd))
		if err != nil && !tc.errExpected {
			t.Fatalf("Error running app: %v", err)
		}
		if err == nil && tc.errExpected {
			t.Fatalf("An error was expected running app: %v", err)
		}

		if actual := buf.String(); actual != tc.expectedStdOutput {
			t.Fatalf("Expected %q, app printed: %q", tc.expectedStdOutput, actual)
		}

		if errValue != 1 && tc.errExpected {
			t.Fatalf("An exit code 1 was expected")
		}

		if errValue == 1 && !tc.errExpected {
			t.Fatalf("An exit code 1 was not expected")
		}

		actual, err := io.ReadAll(ghaOutputFile)
		if err != nil {
			t.Fatalf("Unable to read temporary GHA output file: %v", err)
		}
		if string(actual) != tc.expectedGHAOutput {
			t.Fatalf("Expected %q, GHA output: %q", tc.expectedStdOutput, actual)
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
		cmd               string
		expectedStdOutput string
		expectedGHAOutput string
	}{
		{
			cmd:               fmt.Sprintf("rt -yaml %s is-empty -fail", filepath),
			expectedStdOutput: "false\n",
		},
		{
			cmd:               fmt.Sprintf("rt -gha=1 -yaml %s is-empty -fail", filepath),
			expectedStdOutput: "false\n",
			expectedGHAOutput: "is-empty=false\n",
		},
	} {
		var errValue int
		cli.OsExiter = func(code int) {
			errValue = code
		}

		ghaOutput := testutil.NewGithubOutputWriter(t)

		buf.Reset()
		err = app.Run(strings.Fields(tc.cmd))
		if err != nil {
			t.Fatalf("Error running app: %v", err)
		}

		if actual := buf.String(); actual != tc.expectedStdOutput {
			t.Fatalf("Expected %q, app printed: %q", tc.expectedStdOutput, actual)
		}
		if actual := ghaOutput.Result(t); actual != tc.expectedGHAOutput {
			t.Fatalf("Expected %q, GHA output: %q", tc.expectedStdOutput, actual)
		}

		if errValue != 0 {
			t.Fatalf("Exit code 0 was expected")
		}
	}
}
