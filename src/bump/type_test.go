package bump_test

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/src/bump"
)

func TestBump(t *testing.T) {
	t.Parallel()

	version := "v1.2.3"

	for _, tc := range []struct {
		bt       bump.Type
		expected string
	}{
		{bump.Major, "v2.0.0"},
		{bump.Minor, "v1.3.0"},
		{bump.Patch, "v1.2.4"},
		{bump.None, version},
	} {
		bumped := bump.Bump(semver.MustParse(version), tc.bt)
		if !semver.MustParse(tc.expected).Equal(bumped) {
			t.Fatalf("Bump resulted in %v, expected %v", bumped, tc.expected)
		}
	}
}

func TestBump_Does_Not_Overwrite(t *testing.T) {
	t.Parallel()

	version := semver.MustParse("v1.2.3-beta")
	oldV := version
	newV := bump.Bump(version, bump.Major)

	if !version.Equal(oldV) {
		t.Fatal("Bump modified version in-place!")
	}

	if newV.Equal(oldV) {
		t.Fatal("Unreachable but I do not want compiler to optimize newV out")
	}
}

func TestBumpType_With(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		expected bump.Type
		bumps    []bump.Type
	}{
		{
			expected: bump.Patch,
			bumps: []bump.Type{
				bump.None,
				bump.Patch,
				bump.Patch,
				bump.None,
			},
		},
		{
			expected: bump.Minor,
			bumps: []bump.Type{
				bump.None,
				bump.Patch,
				bump.Patch,
				bump.Minor,
				bump.Patch,
				bump.None,
			},
		},
		{
			expected: bump.Major,
			bumps: []bump.Type{
				bump.None,
				bump.Patch,
				bump.Patch,
				bump.Minor,
				bump.Major,
				bump.Patch,
				bump.None,
			},
		},
	} {
		var finalBump bump.Type
		for _, bump := range tc.bumps {
			finalBump = finalBump.With(bump)
		}

		if finalBump != tc.expected {
			t.Fatalf("Expected bump bump %v, got %v", tc.expected, finalBump)
		}
	}
}

func TestBumpType_Cap(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		expected bump.Type
		bump     bump.Type
		cap      bump.Type
	}{
		// Patcj
		{
			bump:     bump.Patch,
			cap:      bump.Patch,
			expected: bump.Patch,
		},
		{
			bump:     bump.Patch,
			cap:      bump.Minor,
			expected: bump.Patch,
		},
		{
			bump:     bump.Patch,
			cap:      bump.Major,
			expected: bump.Patch,
		},
		// Minor
		{
			bump:     bump.Minor,
			cap:      bump.Patch,
			expected: bump.Patch,
		},
		{
			bump:     bump.Minor,
			cap:      bump.Minor,
			expected: bump.Minor,
		},
		{
			bump:     bump.Minor,
			cap:      bump.Major,
			expected: bump.Minor,
		},
		// Major
		{
			bump:     bump.Major,
			cap:      bump.Patch,
			expected: bump.Patch,
		},
		{
			bump:     bump.Major,
			cap:      bump.Minor,
			expected: bump.Minor,
		},
		{
			bump:     bump.Major,
			cap:      bump.Major,
			expected: bump.Major,
		},
	} {
		actual := tc.bump.Cap(tc.cap)
		if actual != tc.expected {
			t.Fatalf("Expected %v for %v.Cap(%v), got %v", tc.expected, tc.bump, tc.cap, actual)
		}
	}
}

func TestBump_From(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		from     string
		to       string
		expected bump.Type
	}{
		{
			from:     "v1.2.3",
			to:       "v1.2.4",
			expected: bump.Patch,
		},
		{
			from:     "v1.2.3",
			to:       "v1.4.0",
			expected: bump.Minor,
		},
		{
			from:     "v1.2.3",
			to:       "v2.4.0",
			expected: bump.Major,
		},
	} {
		actual := bump.From(semver.MustParse(tc.from), semver.MustParse(tc.to))
		if actual != tc.expected {
			t.Fatalf("Expected %v for %v -> %v, got %v", tc.expected, tc.from, tc.to, actual)
		}
	}
}
