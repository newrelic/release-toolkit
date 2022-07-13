// Package app holds the cli.App that allows running the release toolkit components as command line apps.
package app

import (
	"github.com/newrelic/release-toolkit/app/common"
	"github.com/urfave/cli/v2"
)

// App returns a cli.App with all known commands added to it.
func App() *cli.App {
	return &cli.App{
		Name:  "rt",
		Usage: "Release toolkit",
		Flags: []cli.Flag{
			// -changelog is the command line flag to specify the path to a changelog.yaml file.
			// This flag is common and used by most commands.
			&cli.StringFlag{
				Name:    common.ChangelogFlag,
				EnvVars: common.EnvFor(common.ChangelogFlag),
				Value:   "changelog.yaml",
				Usage:   "Path to the changelog.yaml file",
			},
		},
	}
}
