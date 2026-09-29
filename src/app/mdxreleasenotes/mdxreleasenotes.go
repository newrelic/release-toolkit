package mdxreleasenotes

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/newrelic/release-toolkit/src/app/common"
	"github.com/newrelic/release-toolkit/src/app/gha"
	"github.com/newrelic/release-toolkit/src/mdxreleasenotes"
	"github.com/urfave/cli/v2"
)

const (
	changelogPathFlag = "changelog"
	versionFlag       = "version"
	subjectFlag       = "subject"
	repoFlag          = "repo"
)

const mdxPathOutput = "mdx-path"

// Cmd is the cli.Command object for the release-notes-mdx command.
//
//nolint:gochecknoglobals // We could overengineer this to avoid the global command but I don't think it's worth it.
var Cmd = &cli.Command{
	Name: "release-notes-mdx",
	Usage: "Extracts a version's section from CHANGELOG.md and renders it as an MDX file matching the " +
		"newrelic/docs-website release notes schema.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     subjectFlag,
			EnvVars:  common.EnvFor(subjectFlag),
			Usage:    "Product or component name to stamp in the `subject` frontmatter field (e.g. \"Agent Control\").",
			Required: true,
		},
		&cli.StringFlag{
			Name:     versionFlag,
			EnvVars:  common.EnvFor(versionFlag),
			Usage:    "Version to extract from the changelog, without a `v` prefix (e.g. 1.2.3).",
			Required: true,
		},
		&cli.StringFlag{
			Name:     repoFlag,
			EnvVars:  common.EnvFor(repoFlag),
			Usage:    "GitHub repository in `owner/name` form, used to build the release notes link.",
			Required: true,
		},
		&cli.StringFlag{
			Name:    changelogPathFlag,
			EnvVars: common.EnvFor(changelogPathFlag),
			Usage:   "Path to the source CHANGELOG.md file.",
			Value:   "CHANGELOG.md",
		},
	},
	Action: Run,
}

// Run is a command function which extracts a version's section from a changelog and writes it as MDX.
func Run(cCtx *cli.Context) error {
	gh := gha.NewFromCli(cCtx)

	mdPath := cCtx.String(changelogPathFlag)

	changelogBytes, err := os.ReadFile(mdPath)
	if err != nil {
		return fmt.Errorf("reading changelog file %q: %w", mdPath, err)
	}

	repo := cCtx.String(repoFlag)
	version := cCtx.String(versionFlag)

	name := path.Base(repo)
	dashedVersion := strings.ReplaceAll(version, ".", "-")
	outputPath := fmt.Sprintf("%s-%s.mdx", name, dashedVersion)

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating destination file at %q: %w", outputPath, err)
	}
	defer outputFile.Close()

	r := mdxreleasenotes.Renderer{
		Subject:  cCtx.String(subjectFlag),
		Repo:     repo,
		Sections: mdxreleasenotes.DefaultSections,
	}

	if err := r.Render(outputFile, string(changelogBytes), version); err != nil {
		return fmt.Errorf("rendering release notes mdx: %w", err)
	}

	_, _ = fmt.Fprintln(cCtx.App.Writer, outputPath)
	gh.SetOutput(mdxPathOutput, outputPath)

	return nil
}
