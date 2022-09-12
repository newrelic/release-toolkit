package hack

import "github.com/Masterminds/semver"

// SemverEquals is a function to be used with cmp.Diff, which compare two semver pointers while allowing them to be nil.
// This is required because semver.Version.Compare does not handle nils on its own.
func SemverEquals(va, vb *semver.Version) bool {
	if va == nil || vb == nil {
		return va == nil && vb == nil
	}

	return va.Compare(vb) == 0
}
