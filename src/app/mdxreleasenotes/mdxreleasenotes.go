// Package mdxreleasenotes wires the mdxreleasenotes.Renderer into the rt CLI as the release-notes-mdx command.
package mdxreleasenotes

import (
	"fmt"
	"os"
	"strings"

	"github.com/newrelic/release-toolkit/src/app/common"
	"github.com/newrelic/release-toolkit/src/mdxreleasenotes"
	"github.com/urfave/cli/v2"
)

const (
	markdownPathFlag = "markdown"
	versionFlag      = "version"
	subjectFlag      = "subject"
	repoFlag         = "repo"
	outputFlag       = "output"
)

// Cmd is the cli.Command object for the release-notes-mdx command.
//
//nolint:gochecknoglobals // We could overengineer this to avoid the global command but I don't think it's worth it.
var Cmd = &cli.Command{
	Name:  "release-notes-mdx",
	Usage: "Extracts a version's section from CHANGELOG.md and renders it as a docs-site MDX file.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    markdownPathFlag,
			EnvVars: common.EnvFor(markdownPathFlag),
			Usage:   "Path to the source CHANGELOG.md file.",
			Value:   "CHANGELOG.md",
		},
		&cli.StringFlag{
			Name:     versionFlag,
			EnvVars:  common.EnvFor(versionFlag),
			Usage:    "Version to extract from the changelog, without a `v` prefix (e.g. 1.2.3).",
			Required: true,
		},
		&cli.StringFlag{
			Name:     subjectFlag,
			EnvVars:  common.EnvFor(subjectFlag),
			Usage:    "Product or component name to stamp in the `subject` frontmatter field (e.g. \"Agent Control\").",
			Required: true,
		},
		&cli.StringFlag{
			Name:     repoFlag,
			EnvVars:  common.EnvFor(repoFlag),
			Usage:    "GitHub repository in `owner/name` form, used to build the release notes link.",
			Required: true,
		},
		&cli.StringFlag{
			Name:    outputFlag,
			EnvVars: common.EnvFor(outputFlag),
			Usage:   "Path of the MDX file to write. If omitted, defaults to \"<repo-name>-<version-with-dashes>.mdx\".",
			Value:   "",
		},
	},
	Action: Run,
}

// Run is a command function which extracts a version's section from a markdown changelog and writes it as MDX.
func Run(cCtx *cli.Context) error {
	mdPath := cCtx.String(markdownPathFlag)

	changelogBytes, err := os.ReadFile(mdPath)
	if err != nil {
		return fmt.Errorf("reading changelog file %q: %w", mdPath, err)
	}

	version := cCtx.String(versionFlag)

	outputPath := cCtx.String(outputFlag)
	if outputPath == "" {
		outputPath = defaultOutputPath(cCtx.String(repoFlag), version)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating destination file at %q: %w", outputPath, err)
	}
	defer outputFile.Close()

	r := mdxreleasenotes.Renderer{
		Subject:  cCtx.String(subjectFlag),
		Repo:     cCtx.String(repoFlag),
		Sections: mdxreleasenotes.DefaultSections,
	}

	if err := r.Render(outputFile, string(changelogBytes), version); err != nil {
		return fmt.Errorf("rendering release notes mdx: %w", err)
	}

	fmt.Println(outputPath)

	return nil
}

// defaultOutputPath derives "<repo-name>-<version-with-dashes>.mdx" from a `owner/name` repo slug and a version.
func defaultOutputPath(repo, version string) string {
	name := repo
	if idx := strings.LastIndexByte(repo, '/'); idx != -1 {
		name = repo[idx+1:]
	}

	dashedVersion := strings.ReplaceAll(version, ".", "-")

	return fmt.Sprintf("%s-%s.mdx", name, dashedVersion)
}
