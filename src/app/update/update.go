package update

import (
	"fmt"
	"os"
	"time"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/src/app/common"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/changelog/sources/markdown/merger"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

const (
	markdownPathFlag = "markdown"
	versionFlag      = "version"
	dateFlag         = "date"
)

// Cmd is the cli.Command object for the update-changelog command.
//
//nolint:gochecknoglobals // We could overengineer this to avoid the global command but I don't think it's worth it.
var Cmd = &cli.Command{
	Name:  "update-changelog",
	Usage: "Incorporates a changelog.yaml into a complete CHANGELOG.md.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     markdownPathFlag,
			EnvVars:  common.EnvFor(markdownPathFlag),
			Usage:    "Path to the destination markdown file.",
			Value:    "CHANGELOG.md",
			Required: true,
		},
		&cli.StringFlag{
			Name:    versionFlag,
			EnvVars: common.EnvFor(versionFlag),
			Usage: "Version to stamp in the changelog section header. " +
				"If omitted, no version header will be generated.",
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
	Action: Update,
}

// Update is a command function which loads a changelog.yaml file and merges it into an existing CHANGELOG.md document.
func Update(cCtx *cli.Context) error {
	chPath := cCtx.String(common.YAMLFlag)
	chFile, err := os.Open(chPath)
	if err != nil {
		return fmt.Errorf("opening changelog file %q: %w", chPath, err)
	}

	ch := &changelog.Changelog{}
	err = yaml.NewDecoder(chFile).Decode(ch)
	if err != nil {
		return fmt.Errorf("loading changelog from file: %w", err)
	}

	currentMdPath := cCtx.String(markdownPathFlag)
	newMdPath := currentMdPath + ".new"
	bakMdPath := currentMdPath + ".bak"

	currentMdFile, err := os.Open(currentMdPath)
	if err != nil {
		return fmt.Errorf("opening existing changelog at %q: %w", currentMdPath, err)
	}

	newMdFile, err := os.Create(newMdPath)
	if err != nil {
		return fmt.Errorf("creating destination file at %q: %w", newMdPath, err)
	}

	version, err := semver.NewVersion(cCtx.String(versionFlag))
	if err != nil {
		return fmt.Errorf("parsing version: %w", err)
	}

	mrg := merger.New(ch, version)
	if t := cCtx.Timestamp(dateFlag); t != nil {
		tv := *t
		mrg.ReleasedOn = func() time.Time {
			return tv
		}
	}

	err = mrg.Merge(currentMdFile, newMdFile)
	if err != nil {
		return fmt.Errorf("merging changelog: %w", err)
	}

	_ = os.Rename(currentMdPath, bakMdPath)
	_ = os.Rename(newMdPath, currentMdPath)

	return nil
}
