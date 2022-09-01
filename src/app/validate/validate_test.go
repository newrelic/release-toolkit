package validate_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/newrelic/release-toolkit/app"
)

//nolint:paralleltest // urfave/cli cannot be tested concurrently.
func TestValidate(t *testing.T) {
	for _, tc := range []struct {
		name     string
		md       string
		args     string
		expected string
	}{
		{
			name: "Changelog_With_Two_Errors",
			md: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unreleased

### Important announcement (note)

### Breaking
Support has been removed

### Security
- Fixed a security issue that leaked all data

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should not be included
`),
			args: "--exit-code=0",
			expected: strings.TrimLeft(`
"Important announcement (note)" header found with empty content
"Breaking" header must contain only an itemized list
`, "\n"),
		},
		{
			name: "Valid_Changelog",
			md: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unreleased

### Important announcement (note)
This is a release note

### Breaking
- Support has been removed

### Security
- Fixed a security issue that leaked all data

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should not be included
`),
			args:     "--exit-code=0",
			expected: "",
		},
	} {
		tc := tc
		//nolint:paralleltest // urfave/cli cannot be tested concurrently.
		t.Run(tc.name, func(t *testing.T) {
			tDir := t.TempDir()

			app := app.App()
			buf := &strings.Builder{}
			app.ErrWriter = buf
			app.Writer = buf

			mdPath := path.Join(tDir, "CHANGELOG.md")
			mdFile, err := os.Create(mdPath)
			if err != nil {
				t.Fatalf("Error creating test markdown source: %v", err)
			}
			defer mdFile.Close()

			_, _ = mdFile.WriteString(tc.md)

			err = app.Run(strings.Fields(fmt.Sprintf("rt validate -md %s %s", mdPath, tc.args)))
			if err != nil {
				t.Fatalf("Error running app: %v", err)
			}

			if actual := buf.String(); actual != tc.expected {
				t.Fatalf("Expected:\n%s\n\napp printed:\n%s", tc.expected, actual)
			}
		})
	}
}
