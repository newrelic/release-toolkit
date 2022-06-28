package bumptype_test

import (
	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/bump/bumptype"
	"testing"
)

func TestBump(t *testing.T) {
	t.Parallel()

	version := "v1.2.3-beta"

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
