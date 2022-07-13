package render

import (
	"fmt"
	"os"
	"time"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/app/common"
	"github.com/newrelic/release-toolkit/changelog"
	"github.com/newrelic/release-toolkit/changelog/renderer"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

const (
	mdPathFlag  = "md"
	versionFlag = "version"
	dateFlag    = "date"
)

// Cmd is the cli.Command object for the is-held command.
//nolint:gochecknoglobals // We could overengineer this to avoid the global command but I don't think it's worth it.
var Cmd = &cli.Command{
	Name:  "render-changelog",
	Usage: "Renders a changelog.yaml as a markdown changelog section.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    mdPathFlag,
			EnvVars: common.EnvFor(mdPathFlag),
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
		&cli.StringFlag{
			Name:    dateFlag,
			EnvVars: common.EnvFor(dateFlag),
			Usage: "Date to stamp in the changelog section header, in YYYY-MM-DD format. " +
				"If empty it will be omitted from the header. The literal 'now' can also be specified.",
			Value: "now",
		},
	},
	Action: Render,
}

// Render is a command function which loads a changelog.yaml file from this, and prints to stdout whether it has the
// Held flag set to true.
func Render(cCtx *cli.Context) error {
	chPath := cCtx.String(common.ChangelogFlag)
	chFile, err := os.Open(chPath)
	if err != nil {
		return fmt.Errorf("opening changelog file %q: %w", chPath, err)
	}

	ch := &changelog.Changelog{}
	err = yaml.NewDecoder(chFile).Decode(ch)
	if err != nil {
		return fmt.Errorf("loading changelog from file: %w", err)
	}

	mdPath := cCtx.String(mdPathFlag)
	mdFile, err := os.Create(mdPath)
	if err != nil {
		return fmt.Errorf("creating destination file at %q: %w", mdPath, err)
	}

	rnd := renderer.New(ch)
	err = addDate(&rnd, cCtx.String(dateFlag))
	if err != nil {
		return err
	}

	err = addVersion(&rnd, cCtx.String(versionFlag))
	if err != nil {
		return err
	}

	err = rnd.Render(mdFile)
	if err != nil {
		return fmt.Errorf("rendering changelog: %w", err)
	}

	return nil
}

func addDate(rnd *renderer.Renderer, dateStr string) error {
	switch dateStr {
	case "":
		log.Infof("Generating changelog without a release date")
	case "now":
		rnd.ReleasedOn = time.Now
	default:
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return fmt.Errorf("parsing date %q: %w", dateStr, err)
		}

		rnd.ReleasedOn = func() time.Time {
			return date
		}
	}

	return nil
}

func addVersion(rnd *renderer.Renderer, versionStr string) error {
	if versionStr != "" {
		version, err := semver.NewVersion(versionStr)
		if err != nil {
			return fmt.Errorf("parsing version %q: %w", versionStr, err)
		}
		rnd.Next = version
	}

	return nil
}
