// Package gha is a convenience object to use Workflow commands for Github Actions.
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions
package gha

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/newrelic/release-toolkit/src/app/common"
	"github.com/urfave/cli/v2"
)

var ErrEmptyFileName = errors.New("filename is an empty string")

const (
	GithubOutput = "GITHUB_OUTPUT"
)

// New creates a Github object that will print commands to the specified writer.
func New(writer io.Writer) Github {
	return Github{w: writer}
}

// NewFromCli takes a cli.Context and looks for common.GHAFlag on it. If it is set, it returns a Github object that
// writes to the app's Writer. If it is not, it returns an empty Github object that does not write anything.
func NewFromCli(cCtx *cli.Context) (Github, error) {
	if !cCtx.Bool(common.GHAFlag) {
		return New(io.Discard), nil
	}

	filename := os.Getenv(GithubOutput)

	if filename == "" {
		return Github{}, fmt.Errorf("invalid output file: %w", ErrEmptyFileName)
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644) //nolint:mnd,magicnumber
	if err != nil {
		return Github{}, fmt.Errorf("invalid output file: %w", err)
	}

	return New(file), nil
}

// Github is an object that output workflow commands.
type Github struct {
	w  io.Writer
	mu sync.Mutex
}

// SetOutput outputs the `set-output` command.
// https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions#setting-an-output-parameter
func (g *Github) SetOutput(name string, value interface{}) {
	g.mu.Lock()
	_, _ = fmt.Fprintf(g.w, "%s=%v\n", name, value)
	g.mu.Unlock()
}
