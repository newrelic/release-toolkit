package tag

import (
	"fmt"

	"github.com/Masterminds/semver"
)

type Tag struct {
	Version *semver.Version
	Hash    string
}

// Source is an object that returns a list of (unordered) versions.
// Typical implementations would be a git repository, or simply a hardcoded string.
type Source interface {
	Tags() ([]Tag, error)
}

// Static is a fixed tag that returns itself as the only tag with semver tag and empty hash.
type Static string

// Tags attempts to parse the underlying string as a semver and returns it.
func (ss Static) Tags() ([]Tag, error) {
	version, err := semver.NewVersion(string(ss))
	if err != nil {
		return nil, fmt.Errorf("parsing version from %q: %w", string(ss), err)
	}

	return []Tag{
		{
			Version: version,
		},
	}, nil
}
