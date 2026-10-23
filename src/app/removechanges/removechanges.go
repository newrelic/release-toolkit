package removechanges

import (
	"fmt"

	"github.com/newrelic/release-toolkit/src/app/common"
	"github.com/newrelic/release-toolkit/src/changelog/sources/changes"
	"github.com/urfave/cli/v2"
)

const changesDirFlag = "changes-dir"

//nolint:gochecknoglobals // We could overengineer this to avoid the global command but I don't think it's worth it.
var Cmd = &cli.Command{
	Name:  "remove-changes",
	Usage: "Removes the YAML files, one per change, that hold the unreleased changelog entries, once they have been released",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    changesDirFlag,
			EnvVars: common.EnvFor(changesDirFlag),
			Usage:   "Directory with the YAML files to remove",
			Value:   changes.DefaultDir,
		},
	},
	Action: Remove,
}

func Remove(cCtx *cli.Context) error {
	removed, err := changes.NewSource(cCtx.String(changesDirFlag)).Remove()

	for _, path := range removed {
		_, _ = fmt.Fprintln(cCtx.App.Writer, path)
	}

	if err != nil {
		return fmt.Errorf("removing change files: %w", err)
	}

	return nil
}
