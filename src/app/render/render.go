package render

import (
	"fmt"
	"os"
	"time"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/src/app/common"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/changelog/renderer"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

const (
	markdownPathFlag = "markdown"
	versionFlag      = "version"
	dateFlag         = "date"
)

// Cmd is the cli.Command object for the is-held command.
//
//nolint:gochecknoglobals // We could overengineer this to avoid the global command but I don't think it's worth it.
var Cmd = &cli.Command{
	Name:  "render-changelog",
	Usage: "Renders a changelog.yaml as a markdown changelog section.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    markdownPathFlag,
			EnvVars: common.EnvFor(markdownPathFlag),
			Usage:   "Path to the destination markdown file.",
			Value:   "CHANGELOG.partial.md",
		},
		&cli.StringFlag{
			Name:    versionFlag,
			EnvVars: common.EnvFor(versionFlag),
			Usage: "Version to stamp in the changelog section header. " +
				"If omitted, no version header will be generated",
			Value: "",
		},
		&cli.TimestampFlag{
			Name:    dateFlag,
			EnvVars: common.EnvFor(dateFlag),
			Usage: "Date to stamp in the changelog section header, in YYYY-MM-DD format. " +
				"If empty it will default to the current time (time.Now()).",
			Value:  cli.NewTimestamp(time.Now()),
			Layout: "2006-01-02",
		},
	},
	Action: Render,
}

// Render is a command function which loads a changelog.yaml file from this, and prints to stdout whether it has the
// Held flag set to true.
func Render(cCtx *cli.Context) error {
	chPath := cCtx.String(common.YAMLFlag)
	chFile, err := os.Open(chPath)
	if err != nil {
		return fmt.Errorf("opening changelog yaml file %q: %w", chPath, err)
	}

	ch := &changelog.Changelog{}
	err = yaml.NewDecoder(chFile).Decode(ch)
	if err != nil {
		return fmt.Errorf("loading changelog from file: %w", err)
	}

	mdPath := cCtx.String(markdownPathFlag)
	mdFile, err := os.Create(mdPath)
	if err != nil {
		return fmt.Errorf("creating destination file at %q: %w", mdPath, err)
	}

	rnd := renderer.New(ch)

	if t := cCtx.Timestamp(dateFlag); t != nil {
		tv := *t
		rnd.ReleasedOn = func() time.Time {
			return tv
		}
	}

	if versionStr := cCtx.String(versionFlag); versionStr != "" {
		version, vErr := semver.NewVersion(versionStr)
		if vErr != nil {
			return fmt.Errorf("parsing version %q: %w", versionStr, vErr)
		}
		rnd.Next = version
	}

	err = rnd.Render(mdFile)
	if err != nil {
		return fmt.Errorf("rendering changelog: %w", err)
	}

	return nil
}
