package isheld

import (
	"fmt"
	"os"

	"github.com/newrelic/release-toolkit/src/app/common"
	"github.com/newrelic/release-toolkit/src/app/gha"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

const failFlag = "fail"

const isHeldOutput = "is-held"

// Cmd is the cli.Command object for the is-held command.
//
//nolint:gochecknoglobals // We could overengineer this to avoid the global command but I don't think it's worth it.
var Cmd = &cli.Command{
	Name:  "is-held",
	Usage: "Prints `true' if changelog should be held, `false' otherwise.",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    failFlag,
			EnvVars: common.EnvFor(failFlag),
			Usage:   "If set, command will exit with a code of 1 if changelog should be held.",
			Value:   false,
		},
	},
	Action: IsHeld,
}

// IsHeld is a command function which loads a changelog.yaml file from this, and prints to stdout whether it has the
// Held flag set to true.
func IsHeld(cCtx *cli.Context) error {
	gh := gha.NewFromCli(cCtx)

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

	_, _ = fmt.Fprintf(cCtx.App.Writer, "%v\n", ch.Held)
	gh.SetOutput(isHeldOutput, ch.Held)

	if cCtx.Bool("fail") && ch.Held {
		return cli.Exit("", 1)
	}

	return nil
}
