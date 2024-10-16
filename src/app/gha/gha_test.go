package gha_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/newrelic/release-toolkit/internal/testutil"
	"github.com/newrelic/release-toolkit/src/app/gha"
)

//nolint:paralleltest
func TestGHA_FileIsEmptyByDefault(t *testing.T) {
	ghaOutput := testutil.NewGithubOutputWriter(t)
	if actual := ghaOutput.Result(t); actual != "" {
		t.Fatalf("Expected GHA output is empty")
	}
}

func TestGHA_OutputsAreAppended(t *testing.T) {
	t.Parallel()

	buf := &strings.Builder{}
	buf.WriteString("not-empty=true\nanother-line=true\n")

	gha := gha.New(buf)

	gha.SetOutput("test", 1)
	gha.SetOutput("test", "out")

	expected := strings.TrimSpace(`
not-empty=true
another-line=true
test=1
test=out
	`) + "\n"

	if actual := buf.String(); actual != expected {
		t.Fatalf("Expected:\n%s\n\ngot:\n%s", expected, actual)
	}
}

func TestGHA_LockWritesAndDoesNotMangleOutput(t *testing.T) {
	t.Parallel()

	buf := &strings.Builder{}
	wg := sync.WaitGroup{}

	gha := gha.New(buf)

	for range 5 {
		wg.Add(1)
		go func() {
			gha.SetOutput("test", 1)
			wg.Done()
		}()
	}

	expected := strings.TrimSpace(`
test=1
test=1
test=1
test=1
test=1
	`) + "\n"

	wg.Wait()

	if actual := buf.String(); actual != expected {
		t.Fatalf("Expected:\n%s\n\ngot:\n%s", expected, actual)
	}
}
