package bumptype

import "github.com/Masterminds/semver"

type Type int

const (
	None  = Type(0)
	Patch = Type(1)
	Minor = Type(2)
	Major = Type(3)
)

// Less returns whether the current bump Type is smaller than another one.
func (bt Type) Less(other Type) bool {
	return bt < other
}

// With composes two bump types, returning the largest of the two.
func (bt Type) With(other Type) Type {
	if !bt.Less(other) {
		return bt
	}

	return other
}

// Cap returns the current bumptype if it is smaller or equal than another one, and second otherwise.
// e.g Major.Cap(Minor) returns Minor, and Patch.Cap(Minor) returns Patch.
func (bt Type) Cap(other Type) Type {
	if other.Less(bt) {
		return other
	}

	return bt
}

// From deduces the bump that caused the old version to go to the new.
func From(old, new *semver.Version) Type {
	if new.Major() > old.Major() {
		return Major
	}

	if new.Minor() > old.Minor() {
		return Minor
	}

	if new.Patch() > old.Patch() {
		return Patch
	}

	return None
}

// Bump returns a new version after bumping it according to the specified bump bumptype.
func Bump(version *semver.Version, bt Type) *semver.Version {
	if bt == None {
		return version
	}

	v := *version

	//nolint:exhaustive // case None is handled in the if above, saving a copy operation.
	switch bt {
	case Patch:
		v = version.IncPatch()
	case Minor:
		v = version.IncMinor()
	case Major:
		v = version.IncMajor()
	}

	return &v
}
