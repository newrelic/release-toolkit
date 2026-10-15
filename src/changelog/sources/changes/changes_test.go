package changes_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	cl "github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/changelog/sources/changes"
)

func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

func TestSource_Changelog(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		files    map[string]string
		expected *cl.Changelog
		errText  []string
	}{
		{
			name: "Sorted_By_Name",
			files: map[string]string{
				"b-bugfix.yaml":     "type: bugfix\nmessage: Fixed a bug\n",
				"a-enhancement.yml": "type: enhancement\nmessage: Added a feature\n",
				"c-breaking.yaml":   "type: Breaking changes\nmessage: Removed support\n",
				"d-security.yaml":   "type: \"🛡️ Security notices\"\nmessage: Fixed a leak\n",
				"e-dependency.yaml": "type: dependencies\nmessage: Updated foo\n",
				"f-spaces.yaml":     "type: ' Bug Fixes '\nmessage: '  Trimmed  '\n",
			},
			expected: &cl.Changelog{Changes: []cl.Entry{
				{Type: cl.TypeEnhancement, Message: "Added a feature"},
				{Type: cl.TypeBugfix, Message: "Fixed a bug"},
				{Type: cl.TypeBreaking, Message: "Removed support"},
				{Type: cl.TypeSecurity, Message: "Fixed a leak"},
				{Type: cl.TypeDependency, Message: "Updated foo"},
				{Type: cl.TypeBugfix, Message: "Trimmed"},
			}},
		},
		{
			name: "Skips_Files_That_Are_Not_Changes",
			files: map[string]string{
				"change.yaml":        "type: bugfix\nmessage: Fixed a bug\n",
				"TEMPLATE.yaml":      "type:\nmessage:\n",
				"template.yml":       "type:\nmessage:\n",
				".hidden.yaml":       "type: bugfix\nmessage: hidden\n",
				"_ignored.yaml":      "type: bugfix\nmessage: ignored\n",
				"README.md":          "# not a change file\n",
				"notes.txt":          "type: bugfix\nmessage: not yaml\n",
				"nested/deeper.yaml": "type: bugfix\nmessage: nested\n",
			},
			expected: &cl.Changelog{Changes: []cl.Entry{
				{Type: cl.TypeBugfix, Message: "Fixed a bug"},
			}},
		},
		{
			name: "Invalid_Changes_Fail_Listing_All",
			files: map[string]string{
				"typo.yaml": "type: bugfix\nmsg: Fixed a bug\n",
				"bad.yaml":  "type: [unterminated\n",
				"ok.yaml":   "type: bugfix\nmessage: Fixed a bug\n",
			},
			errText: []string{"bad.yaml", "typo.yaml"},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual, err := changes.NewSource(writeFiles(t, tc.files)).Changelog()
			for _, text := range tc.errText {
				if err == nil || !strings.Contains(err.Error(), text) {
					t.Fatalf("Expected an error including %q, got: %v", text, err)
				}
			}

			if len(tc.errText) == 0 && err != nil {
				t.Fatalf("Error loading change files: %v", err)
			}

			if diff := cmp.Diff(tc.expected, actual); diff != "" {
				t.Fatalf("Changes are not as expected: %s", diff)
			}
		})
	}
}

func TestSource_Missing_Directory_Has_No_Changes(t *testing.T) {
	t.Parallel()

	source := changes.NewSource(filepath.Join(t.TempDir(), "does-not-exist"))

	actual, err := source.Changelog()
	if err != nil {
		t.Fatalf("Error loading change files: %v", err)
	}

	if !actual.Empty() {
		t.Fatalf("Expected an empty changelog, got: %+v", actual)
	}

	if errs := source.Validate(); len(errs) != 0 {
		t.Fatalf("Expected no errors, got: %v", errs)
	}
}

func TestSource_Validate(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		content  string
		expected error
		errText  string
	}{
		{name: "Empty_file", content: "", expected: changes.ErrEmptyChange},
		{name: "No_type", content: "message: Fixed a bug\n", expected: changes.ErrNoType},
		{name: "Blank_type", content: "type: '  '\nmessage: Fixed a bug\n", expected: changes.ErrNoType},
		{name: "Unknown_type", content: "type: feature\nmessage: Fixed a bug\n", expected: changes.ErrUnknownType, errText: "bugfix"},
		{name: "No_message", content: "type: bugfix\n", expected: changes.ErrNoMessage},
		{name: "Blank_message", content: "type: bugfix\nmessage: '  '\n", expected: changes.ErrNoMessage},
		{name: "Multiline_message", content: "type: bugfix\nmessage: |\n  one\n  two\n", expected: changes.ErrMultilineMessage},
		{name: "Unknown_field", content: "type: bugfix\nmsg: Fixed a bug\n", errText: "field msg not found"},
		{name: "Meta_field", content: "type: bugfix\nmessage: Fixed\nauthor: octocat\n", errText: "field author not found"},
		{name: "Bad_yaml", content: "type: [unterminated\n", errText: "parsing change file"},
		{name: "Valid", content: "type: bugfix\nmessage: Fixed a bug\n"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			errs := changes.NewSource(writeFiles(t, map[string]string{"change.yaml": tc.content})).Validate()
			if tc.expected == nil && tc.errText == "" {
				if len(errs) != 0 {
					t.Fatalf("Expected no errors, got: %v", errs)
				}

				return
			}

			if len(errs) != 1 || !strings.Contains(errs[0].Error(), "change.yaml") {
				t.Fatalf("Expected a single error naming the file, got: %v", errs)
			}

			if tc.expected != nil && !errors.Is(errs[0], tc.expected) {
				t.Fatalf("Expected %q, got %q", tc.expected, errs[0])
			}

			if !strings.Contains(errs[0].Error(), tc.errText) {
				t.Fatalf("Expected the error to include %q, got %q", tc.errText, errs[0])
			}
		})
	}
}

func TestSource_Remove_Only_Deletes_Changes(t *testing.T) {
	t.Parallel()

	dir := writeFiles(t, map[string]string{
		"one.yaml":      "type: bugfix\nmessage: One\n",
		"two.yml":       "type: bugfix\nmessage: Two\n",
		"TEMPLATE.yaml": "type:\nmessage:\n",
		"README.md":     "keep me\n",
		"nested/x.yaml": "type: bugfix\nmessage: Nested\n",
	})

	removed, err := changes.NewSource(dir).Remove()
	if err != nil {
		t.Fatalf("Error removing change files: %v", err)
	}

	if diff := cmp.Diff([]string{filepath.Join(dir, "one.yaml"), filepath.Join(dir, "two.yml")}, removed); diff != "" {
		t.Fatalf("Removed change files are not as expected: %s", diff)
	}

	for _, name := range []string{"TEMPLATE.yaml", "README.md", "nested/x.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("Expected %q to be kept: %v", name, err)
		}
	}

	for _, name := range []string{"one.yaml", "two.yml"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("Expected %q to be removed, got: %v", name, err)
		}
	}
}
