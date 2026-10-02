package validate

import (
	"bytes"
	"fmt"
	"os"

	"github.com/newrelic/release-toolkit/src/app/common"
	"github.com/newrelic/release-toolkit/src/app/gha"
	"github.com/newrelic/release-toolkit/src/changelog/sources/markdown"
	"github.com/urfave/cli/v2"
)

const (
	markdownPathFlag         = "markdown"
	previousMarkdownPathFlag = "previous-markdown"
	exitCodeFlag             = "exit-code"
	validOutput              = "valid"
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
		&cli.StringFlag{
			Name:    previousMarkdownPathFlag,
			EnvVars: common.EnvFor(previousMarkdownPathFlag),
			Usage: "Path to the previous version of the changelog md file. If set, already released " +
				"versions must not be modified compared to this version.",
			Value: "",
		},
	},
	Action: Validate,
}

// Validate is a command function which loads a changelog.md file, and prints to stderr
// all the errors found.
func Validate(cCtx *cli.Context) error {
	gh := gha.NewFromCli(cCtx)

	mdPath := cCtx.String(markdownPathFlag)
	mdContent, err := os.ReadFile(mdPath)
	if err != nil {
		return fmt.Errorf("opening changelog file %q: %w", mdPath, err)
	}

	validator, err := markdown.NewValidator(bytes.NewReader(mdContent))
	if err != nil {
		return fmt.Errorf("creating validator: %w", err)
	}

	errs := validator.Validate()

	if prevPath := cCtx.String(previousMarkdownPathFlag); prevPath != "" {
		prevContent, errRead := os.ReadFile(prevPath)
		if errRead != nil {
			return fmt.Errorf("opening previous changelog file %q: %w", prevPath, errRead)
		}

		errs = append(errs, validator.ValidateDiff(string(prevContent), string(mdContent))...)
	}

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
