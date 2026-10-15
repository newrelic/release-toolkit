package validatechanges_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/newrelic/release-toolkit/src/app"
)

//nolint:paralleltest // urfave/cli cannot be tested concurrently.
func TestValidateChanges(t *testing.T) {
	for _, tc := range []struct {
		name        string
		ghaArg      string
		files       map[string]string
		args        string
		expectedErr string
		expectedGha string
	}{
		{
			name:        "Valid_Changes",
			files:       map[string]string{"a.yaml": "type: bugfix\nmessage: Fixed\n"},
			expectedErr: "",
			expectedGha: "",
		},
		{
			name:        "Valid_Changes_GH_Action",
			ghaArg:      "-gha=1",
			files:       map[string]string{"a.yaml": "type: bugfix\nmessage: Fixed\n"},
			expectedErr: "",
			expectedGha: "::set-output name=valid::true\n",
		},
		{
			name:        "No_Changes",
			files:       map[string]string{},
			expectedErr: "",
			expectedGha: "",
		},
		{
			name: "Changes_With_Two_Errors",
			args: "--exit-code=0",
			files: map[string]string{
				"a.yaml": "type: bugfix\nmessage: Fixed\n",
				"b.yaml": "type: feature\nmessage: Added\n",
				"c.yaml": "type: bugfix\n",
			},
			expectedErr: strings.TrimLeft(`
b.yaml: unknown type "feature", use one of: breaking, security, enhancement, bugfix, dependency
c.yaml: message is required
`, "\n"),
			expectedGha: "",
		},
		{
			name:        "Changes_With_Errors_GH_Action",
			ghaArg:      "-gha=1",
			args:        "--exit-code=0",
			files:       map[string]string{"a.yaml": "type: bugfix\n"},
			expectedErr: "a.yaml: message is required\n",
			expectedGha: "::set-output name=valid::false\n",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tDir := t.TempDir()

			app := app.App()
			bufErr := &strings.Builder{}
			buf := &strings.Builder{}
			app.ErrWriter = bufErr
			app.Writer = buf

			for name, content := range tc.files {
				if err := os.WriteFile(path.Join(tDir, name), []byte(content), 0o600); err != nil {
					t.Fatalf("Error creating change file: %v", err)
				}
			}

			err := app.Run(strings.Fields(fmt.Sprintf("rt %s validate-changes -changes-dir %s %s", tc.ghaArg, tDir, tc.args)))
			if err != nil {
				t.Fatalf("Error running app: %v", err)
			}

			if actual := strings.ReplaceAll(bufErr.String(), tDir+"/", ""); actual != tc.expectedErr {
				t.Fatalf("Expected:\n%s\n\napp printed:\n%s", tc.expectedErr, actual)
			}

			if actual := buf.String(); actual != tc.expectedGha {
				t.Fatalf("Expected:\n%s\n\napp printed:\n%s", tc.expectedGha, actual)
			}
		})
	}
}
