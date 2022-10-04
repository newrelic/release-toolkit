// Package app holds the cli.App that allows running the release toolkit components as command line apps.
package app

import (
	"github.com/newrelic/release-toolkit/src/app/common"
	"github.com/newrelic/release-toolkit/src/app/generate"
	"github.com/newrelic/release-toolkit/src/app/isheld"
	"github.com/newrelic/release-toolkit/src/app/link"
	"github.com/newrelic/release-toolkit/src/app/nextversion"
	"github.com/newrelic/release-toolkit/src/app/render"
	"github.com/newrelic/release-toolkit/src/app/update"
	"github.com/newrelic/release-toolkit/src/app/validate"
	"github.com/urfave/cli/v2"
)

// App returns a cli.App with all known commands added to it.
func App() *cli.App {
	return &cli.App{
		Name:  "rt",
		Usage: "Release toolkit",
		Flags: []cli.Flag{
			// -yaml is the command line flag to specify the path to a changelog.yaml file.
			// This flag is common and used by most commands.
			&cli.StringFlag{
				Name:    common.YAMLFlag,
				EnvVars: common.EnvFor(common.YAMLFlag),
				Value:   "changelog.yaml",
				Usage:   "Path to the changelog.yaml file",
			},
			// -gha tells commands to output workflow commands as understood by Github Actions.
			&cli.BoolFlag{
				Name:    common.GHAFlag,
				EnvVars: []string{common.GHAEnv},
				Value:   false,
				Usage:   "Set to true to echo Workflow commands to stdout using the Github actions syntax",
			},
		},
		Commands: []*cli.Command{
			isheld.Cmd,
			render.Cmd,
			nextversion.Cmd,
			generate.Cmd,
			update.Cmd,
			validate.Cmd,
			link.Cmd,
		},
	}
}
