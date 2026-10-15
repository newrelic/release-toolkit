package removechanges_test

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/newrelic/release-toolkit/src/app"
)

//nolint:paralleltest // urfave/cli cannot be tested concurrently.
func TestRemoveChanges(t *testing.T) {
	for _, tc := range []struct {
		name     string
		files    map[string]string
		dir      string
		expected string
		removed  []string
		kept     []string
	}{
		{
			name: "Removes_Only_Changes",
			files: map[string]string{
				"a.yaml":        "type: bugfix\nmessage: A\n",
				"b.yml":         "type: bugfix\nmessage: B\n",
				"TEMPLATE.yaml": "type:\nmessage:\n",
				"README.md":     "keep me\n",
			},
			expected: "a.yaml\nb.yml\n",
			removed:  []string{"a.yaml", "b.yml"},
			kept:     []string{"TEMPLATE.yaml", "README.md"},
		},
		{
			name:     "Missing_Directory",
			dir:      "does-not-exist",
			expected: "",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tDir := t.TempDir()

			app := app.App()
			buf := &strings.Builder{}
			app.Writer = buf

			for name, content := range tc.files {
				if err := os.WriteFile(path.Join(tDir, name), []byte(content), 0o600); err != nil {
					t.Fatalf("Error creating change file: %v", err)
				}
			}

			err := app.Run(strings.Fields(fmt.Sprintf("rt remove-changes -changes-dir %s", path.Join(tDir, tc.dir))))
			if err != nil {
				t.Fatalf("Error running app: %v", err)
			}

			if actual := strings.ReplaceAll(buf.String(), tDir+"/", ""); actual != tc.expected {
				t.Fatalf("Expected %q, app printed: %q", tc.expected, actual)
			}

			for _, name := range tc.removed {
				if _, err := os.Stat(path.Join(tDir, name)); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("Expected %q to be removed, got: %v", name, err)
				}
			}

			for _, name := range tc.kept {
				if _, err := os.Stat(path.Join(tDir, name)); err != nil {
					t.Fatalf("Expected %q to be kept: %v", name, err)
				}
			}
		})
	}
}
