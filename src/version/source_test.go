package version_test

import (
	"strings"
	"testing"

	"github.com/newrelic/release-toolkit/src/version"
)

func TestStaticSource(t *testing.T) {
	t.Parallel()

	tagStr := "v1.2.3-beta"

	ss := version.Static(tagStr)
	tags, err := ss.Versions()
	if err != nil {
		t.Fatal(err)
	}

	if len(tags) == 0 {
		t.Fatal("static source returned 0 tags")
	}

	unprefixedTag := strings.TrimPrefix(tagStr, "v")
	if rt := tags[0].String(); rt != unprefixedTag {
		t.Fatalf("unexpected tag %q returned, expected %q", rt, tagStr)
	}
}
