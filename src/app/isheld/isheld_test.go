package isheld_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/newrelic/release-toolkit/src/app"
)

//nolint:paralleltest
func TestIsHeld(t *testing.T) {
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
held: true
changes:
  - type: breaking
    message: this is broken
`)

	for _, tc := range []struct {
		cmd      string
		expected string
	}{
		{
			cmd:      fmt.Sprintf("rt -yaml %s is-held", filepath),
			expected: "true\n",
		},
		{
			cmd:      fmt.Sprintf("rt -gha=1 -yaml %s is-held", filepath),
			expected: "true\n::set-output name=is-held::true\n",
		},
	} {
		buf.Reset()
		err = app.Run(strings.Fields(tc.cmd))
		if err != nil {
			t.Fatalf("Error running app: %v", err)
		}

		if actual := buf.String(); actual != tc.expected {
			t.Fatalf("Expected %q, app printed: %q", tc.expected, actual)
		}
	}
}
