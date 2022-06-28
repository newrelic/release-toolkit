package tag

import (
	"fmt"

	"github.com/Masterminds/semver"
)

// Source is an object that returns a list of (unordered) versions.
// Typical implementations would be a git repository, or simply a hardcoded string.
type Source interface {
	Tags() ([]*semver.Version, error)
}

// Static is a fixed string that returns itself as the only semver tag.
type Static string

// Tags attempts to parse the underlying string as a semver and returns it.
func (ss Static) Tags() ([]*semver.Version, error) {
	version, err := semver.NewVersion(string(ss))
	if err != nil {
		return nil, fmt.Errorf("parsing version from %q: %w", string(ss), err)
	}

	return []*semver.Version{version}, nil
}
