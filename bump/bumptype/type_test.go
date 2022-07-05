package bumptype_test

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/bump/bumptype"
)

func TestBump(t *testing.T) {
	t.Parallel()

	version := "v1.2.3"

	for _, tc := range []struct {
		bt       bumptype.Type
		expected string
	}{
		{bumptype.Major, "v2.0.0"},
		{bumptype.Minor, "v1.3.0"},
		{bumptype.Patch, "v1.2.4"},
		{bumptype.None, version},
	} {
		bumped := bumptype.Bump(semver.MustParse(version), tc.bt)
		if !semver.MustParse(tc.expected).Equal(bumped) {
			t.Fatalf("Bump resulted in %v, expected %v", bumped, tc.expected)
		}
	}
}

func TestBump_Does_Not_Overwrite(t *testing.T) {
	t.Parallel()

	version := semver.MustParse("v1.2.3-beta")
	oldV := version
	newV := bumptype.Bump(version, bumptype.Major)

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
		expected bumptype.Type
		bumps    []bumptype.Type
	}{
		{
			expected: bumptype.Patch,
			bumps: []bumptype.Type{
				bumptype.None,
				bumptype.Patch,
				bumptype.Patch,
				bumptype.None,
			},
		},
		{
			expected: bumptype.Minor,
			bumps: []bumptype.Type{
				bumptype.None,
				bumptype.Patch,
				bumptype.Patch,
				bumptype.Minor,
				bumptype.Patch,
				bumptype.None,
			},
		},
		{
			expected: bumptype.Major,
			bumps: []bumptype.Type{
				bumptype.None,
				bumptype.Patch,
				bumptype.Patch,
				bumptype.Minor,
				bumptype.Major,
				bumptype.Patch,
				bumptype.None,
			},
		},
	} {
		var finalBump bumptype.Type
		for _, bump := range tc.bumps {
			finalBump = finalBump.With(bump)
		}

		if finalBump != tc.expected {
			t.Fatalf("Expected bump bumptype %v, got %v", tc.expected, finalBump)
		}
	}
}

func TestBumpType_Cap(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		expected bumptype.Type
		bump     bumptype.Type
		cap      bumptype.Type
	}{
		// Patcj
		{
			bump:     bumptype.Patch,
			cap:      bumptype.Patch,
			expected: bumptype.Patch,
		},
		{
			bump:     bumptype.Patch,
			cap:      bumptype.Minor,
			expected: bumptype.Patch,
		},
		{
			bump:     bumptype.Patch,
			cap:      bumptype.Major,
			expected: bumptype.Patch,
		},
		// Minor
		{
			bump:     bumptype.Minor,
			cap:      bumptype.Patch,
			expected: bumptype.Patch,
		},
		{
			bump:     bumptype.Minor,
			cap:      bumptype.Minor,
			expected: bumptype.Minor,
		},
		{
			bump:     bumptype.Minor,
			cap:      bumptype.Major,
			expected: bumptype.Minor,
		},
		// Major
		{
			bump:     bumptype.Major,
			cap:      bumptype.Patch,
			expected: bumptype.Patch,
		},
		{
			bump:     bumptype.Major,
			cap:      bumptype.Minor,
			expected: bumptype.Minor,
		},
		{
			bump:     bumptype.Major,
			cap:      bumptype.Major,
			expected: bumptype.Major,
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
		expected bumptype.Type
	}{
		{
			from:     "v1.2.3",
			to:       "v1.2.4",
			expected: bumptype.Patch,
		},
		{
			from:     "v1.2.3",
			to:       "v1.4.0",
			expected: bumptype.Minor,
		},
		{
			from:     "v1.2.3",
			to:       "v2.4.0",
			expected: bumptype.Major,
		},
	} {
		actual := bumptype.From(semver.MustParse(tc.from), semver.MustParse(tc.to))
		if actual != tc.expected {
			t.Fatalf("Expected %v for %v -> %v, got %v", tc.expected, tc.from, tc.to, actual)
		}
	}
}
