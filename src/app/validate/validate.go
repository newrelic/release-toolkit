package validate

import (
	"fmt"
	"os"

	"github.com/newrelic/release-toolkit/src/app/common"
	"github.com/newrelic/release-toolkit/src/app/gha"
	"github.com/newrelic/release-toolkit/src/changelog/sources/markdown"
	"github.com/urfave/cli/v2"
)

const (
	markdownPathFlag = "markdown"
	exitCodeFlag     = "exit-code"
	validOutput      = "valid"
)

// Cmd is the cli.Command object for the validate-markdown command.
//
//nolint:gochecknoglobals // We could overengineer this to avoid the global command but I don't think it's worth it.
var Cmd = &cli.Command{
	Name:  "validate-markdown",
	Usage: "Validates a changelog in markdown format and prints errors if the changelog is invalid",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    markdownPathFlag,
			EnvVars: common.EnvFor(markdownPathFlag),
			Usage:   "Validate specified changelog md file.",
			Value:   "CHANGELOG.md",
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

// Validate is a command function which loads a changelog.md file, and prints to stderr
// all the errors found.
func Validate(cCtx *cli.Context) error {
	gh, err := gha.NewFromCli(cCtx)
	if err != nil {
		return fmt.Errorf("creating github client: %w", err)
	}

	mdPath := cCtx.String(markdownPathFlag)
	chFile, err := os.Open(mdPath)
	if err != nil {
		return fmt.Errorf("opening changelog file %q: %w", mdPath, err)
	}

	validator, err := markdown.NewValidator(chFile)
	if err != nil {
		return fmt.Errorf("creating validator: %w", err)
	}

	errs := validator.Validate()

	for _, err := range errs {
		_, _ = fmt.Fprintln(cCtx.App.ErrWriter, err)
	}
	gh.SetOutput(validOutput, len(errs) == 0)

	exitCode := cCtx.Int(exitCodeFlag)
	if len(errs) > 0 && exitCode != 0 {
		return cli.Exit("invalid changelog", exitCode)
	}

	return nil
}
