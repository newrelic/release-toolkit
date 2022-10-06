// Package gha is a convenience object to use Workflow commands for Github Actions.
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions
package gha

import (
	"fmt"
	"io"

	"github.com/newrelic/release-toolkit/src/app/common"
	"github.com/urfave/cli/v2"
)

// New creates a Github object that will print commands to the specified writer.
func New(writer io.Writer) Github {
	return Github{w: writer}
}

// NewFromCli takes a cli.Context and looks for common.GHAFlag on it. If it is set, it returns a Github object that
// writes to the app's Writer. If it is not, it returns an empty Github object that does not write anything.
func NewFromCli(cCtx *cli.Context) Github {
	if cCtx.Bool(common.GHAFlag) {
		return New(cCtx.App.Writer)
	}

	return New(io.Discard)
}

// Github is an object that output workflow commands.
type Github struct {
	w io.Writer
}

// SetOutput outputs the `set-output` command.
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-output-parameter
func (g Github) SetOutput(name string, value interface{}) {
	_, _ = fmt.Fprintf(g.w, "::set-output name=%s::%v\n", name, value)
}
