package markdown

import (
	"errors"
	"regexp"
	"strings"
)

// l2HeaderRegex matches an ATX L2 (##) markdown header line.
var l2HeaderRegex = regexp.MustCompile(`^##\s+(.*?)\s*$`)

// ErrChangesInReleasedVersions is returned when newContent modifies the already-released sections of oldContent,
// i.e. anything after the Unreleased section.
var ErrChangesInReleasedVersions = errors.New("changelog has changes in already released versions")

// ValidateDiff ensures that newContent does not modify the already-released sections of oldContent, i.e. that
// nothing after the Unreleased section was added, removed, or modified.
func (v *Validator) ValidateDiff(oldContent, newContent string) []error {
	oldAfter, oldFound := releasedVersions(oldContent)
	newAfter, newFound := releasedVersions(newContent)

	if !oldFound || !newFound {
		// No Unreleased section found, ErrNoUnreleasedL2Header will already be reported by Validate.
		return nil
	}

	if oldAfter != newAfter {
		return []error{ErrChangesInReleasedVersions}
	}

	return nil
}

// releasedVersions returns the part of content after its Unreleased L2 section, i.e. the already released
// versions. found is false if content has no Unreleased L2 header.
func releasedVersions(content string) (string, bool) {
	lines := strings.Split(content, "\n")

	found := false
	end := len(lines)

	for i, l := range lines {
		match := l2HeaderRegex.FindStringSubmatch(l)
		if match == nil {
			continue
		}

		if !found {
			found = strings.ToLower(match[1]) == unreleasedHeader
			continue
		}

		// We already found the Unreleased header, and this is the next L2 header: released versions start here.
		end = i
		break
	}

	if !found {
		return "", false
	}

	return strings.Join(lines[end:], "\n"), true
}
