package validatechanges

import (
	"fmt"

	"github.com/newrelic/release-toolkit/src/app/common"
	"github.com/newrelic/release-toolkit/src/app/gha"
	"github.com/newrelic/release-toolkit/src/changelog/sources/changes"
	"github.com/urfave/cli/v2"
)

const (
	changesDirFlag = "changes-dir"
	exitCodeFlag   = "exit-code"
	validOutput    = "valid"
)

//nolint:gochecknoglobals // We could overengineer this to avoid the global command but I don't think it's worth it.
var Cmd = &cli.Command{
	Name:  "validate-changes",
	Usage: "Validates the YAML files, one per change, that hold the unreleased changelog entries and prints errors if any is invalid",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    changesDirFlag,
			EnvVars: common.EnvFor(changesDirFlag),
			Usage:   "Directory with the YAML files to validate",
			Value:   changes.DefaultDir,
		},
		&cli.IntFlag{
			Name:    exitCodeFlag,
			EnvVars: common.EnvFor(exitCodeFlag),
			Usage:   "Exit code when errors are found",
			Value:   1,
		},
	},
	Action: Validate,
}

func Validate(cCtx *cli.Context) error {
	gh := gha.NewFromCli(cCtx)

	errs := changes.NewSource(cCtx.String(changesDirFlag)).Validate()
	for _, err := range errs {
		_, _ = fmt.Fprintln(cCtx.App.ErrWriter, err)
	}

	gh.SetOutput(validOutput, len(errs) == 0)

	exitCode := cCtx.Int(exitCodeFlag)
	if len(errs) > 0 && exitCode != 0 {
		return cli.Exit("invalid change files", exitCode)
	}

	return nil
}
