package validate_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/newrelic/release-toolkit/src/app"
)

//nolint:paralleltest,funlen // urfave/cli cannot be tested concurrently.
func TestValidate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		ghaArg      string
		md          string
		previousMd  string
		args        string
		expectedErr string
		expectedGha string
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
			expectedErr: strings.TrimLeft(`
"Important announcement (note)" header found with empty content
"Breaking" header must contain only an itemized list
`, "\n"),
		},
		{
			name:   "Changelog_With_Two_Errors_GH_Action",
			ghaArg: "-gha=1",
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
			expectedErr: strings.TrimLeft(`
"Important announcement (note)" header found with empty content
"Breaking" header must contain only an itemized list
`, "\n"),
			expectedGha: "::set-output name=valid::false\n",
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
			args:        "--exit-code=0",
			expectedErr: "",
		},
		{
			name:   "Valid_Changelog_GH_Action",
			ghaArg: "-gha=1",
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
			args:        "--exit-code=0",
			expectedErr: "",
			expectedGha: "::set-output name=valid::true\n",
		},
		{
			name: "Invalid_Changelog_Only_Notes",
			md: strings.TrimSpace(`
# Changelog
This is based on blah blah blah

## Unreleased

### Important announcement (note)
This is a release note

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should not be included
`),
			args: "--exit-code=0",
			expectedErr: strings.TrimLeft(`
unreleased changelog can't only contain notes
`, "\n"),
		},
		{
			name: "Previous_Markdown_Line_Added_In_Unreleased",
			previousMd: strings.TrimSpace(`
# Changelog

## Unreleased

### Breaking
- Support has been removed

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should be preserved
`),
			md: strings.TrimSpace(`
# Changelog

## Unreleased

### Breaking
- Support has been removed
- Another one

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should be preserved
`),
			args:        "--exit-code=0",
			expectedErr: "",
		},
		{
			name: "Previous_Markdown_Line_Added_Before_Unreleased",
			previousMd: strings.TrimSpace(`
# Changelog

## Unreleased

### Breaking
- Support has been removed

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should be preserved
`),
			md: strings.TrimSpace(`
# Changelog
Sneaked in here

## Unreleased

### Breaking
- Support has been removed

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should be preserved
`),
			args:        "--exit-code=0",
			expectedErr: "",
		},
		{
			name: "Previous_Markdown_Line_Added_In_Released_Version",
			previousMd: strings.TrimSpace(`
# Changelog

## Unreleased

### Breaking
- Support has been removed

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should be preserved
`),
			md: strings.TrimSpace(`
# Changelog

## Unreleased

### Breaking
- Support has been removed

## v1.2.3 - 20YY-DD-MM

### Enhancements
- This is in the past and should be preserved
- Sneaked in
`),
			args: "--exit-code=0",
			expectedErr: strings.TrimLeft(`
changelog has changes in already released versions
`, "\n"),
		},
	} {
		tc := tc
		//nolint:paralleltest // urfave/cli cannot be tested concurrently.
		t.Run(tc.name, func(t *testing.T) {
			tDir := t.TempDir()

			app := app.App()
			bufErr := &strings.Builder{}
			buf := &strings.Builder{}
			app.ErrWriter = bufErr
			app.Writer = buf

			mdPath := path.Join(tDir, "CHANGELOG.md")
			mdFile, err := os.Create(mdPath)
			if err != nil {
				t.Fatalf("Error creating test markdown source: %v", err)
			}
			defer mdFile.Close()

			_, _ = mdFile.WriteString(tc.md)

			if tc.previousMd != "" {
				prevPath := path.Join(tDir, "PREVIOUS_CHANGELOG.md")
				prevFile, errPrev := os.Create(prevPath)
				if errPrev != nil {
					t.Fatalf("Error creating previous markdown source: %v", errPrev)
				}
				defer prevFile.Close()

				_, _ = prevFile.WriteString(tc.previousMd)

				tc.args = fmt.Sprintf("%s -previous-markdown %s", tc.args, prevPath)
			}

			err = app.Run(strings.Fields(fmt.Sprintf("rt %s validate-markdown -markdown %s %s", tc.ghaArg, mdPath, tc.args)))
			if err != nil {
				t.Fatalf("Error running app: %v", err)
			}

			if actual := bufErr.String(); actual != tc.expectedErr {
				t.Fatalf("Expected:\n%s\n\napp printed:\n%s", tc.expectedErr, actual)
			}

			if actual := buf.String(); actual != tc.expectedGha {
				t.Fatalf("Expected:\n%s\n\napp printed:\n%s", tc.expectedGha, actual)
			}
		})
	}
}
