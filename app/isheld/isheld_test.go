package isheld_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/newrelic/release-toolkit/app"
)

func TestIsHeld(t *testing.T) {
	t.Parallel()

	app := app.App()

	buf := &strings.Builder{}
	app.Writer = buf

	dir := t.TempDir()
	filepath := path.Join(dir, "changelog.yaml")
	file, err := os.Create(filepath)
	if err != nil {
		t.Fatalf("Error creating yaml for test: %v", err)
	}

	_, _ = file.WriteString(`
held: true
changes:
  - type: breaking
    message: this is broken
`)

	err = app.Run(strings.Fields(fmt.Sprintf("rt -changelog %s is-held", filepath)))
	if err != nil {
		t.Fatalf("Error running app: %v", err)
	}

	if actual := buf.String(); actual != "true" {
		t.Fatalf("Expected 'true', app printed:\n%s", actual)
	}
}
