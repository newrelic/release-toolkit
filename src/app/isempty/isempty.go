package isempty

import (
	"fmt"
	"os"

	"github.com/newrelic/release-toolkit/src/app/common"
	"github.com/newrelic/release-toolkit/src/app/gha"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

const isEmptyOutput = "is-empty"

const failFlag = "fail"

// Cmd is the cli.Command object for the is-empty command.
//
//nolint:gochecknoglobals
var Cmd = &cli.Command{
	Name:  "is-empty",
	Usage: "Outputs whether automated releases are not needed since changelog is empty",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    failFlag,
			EnvVars: common.EnvFor(failFlag),
			Usage:   "If set, command will exit with a code of 1 if changelog is empty.",
			Value:   false,
		},
	}, Action: IsEmpty,
}

// IsEmpty is a command function which loads a changelog.yaml file, and prints to stdout whether it is empty or not.
func IsEmpty(cCtx *cli.Context) error {
	gh, err := gha.NewFromCli(cCtx)
	if err != nil {
		return fmt.Errorf("creating github client: %w", err)
	}

	chPath := cCtx.String(common.YAMLFlag)
	chFile, err := os.Open(chPath)
	if err != nil {
		return fmt.Errorf("opening changelog.yml file %q: %w", chPath, err)
	}

	ch := changelog.Changelog{}
	err = yaml.NewDecoder(chFile).Decode(&ch)
	if err != nil {
		return fmt.Errorf("loading changelog from file: %w", err)
	}

	chEmpty := ch.Empty()
	_, _ = fmt.Fprintf(cCtx.App.Writer, "%v\n", chEmpty)

	gh.SetOutput(isEmptyOutput, chEmpty)
	if cCtx.Bool("fail") && chEmpty {
		return cli.Exit("changelog is empty", 1)
	}

	return nil
}
