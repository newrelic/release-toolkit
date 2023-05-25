package generate

import (
	"errors"
	"fmt"
	"os"

	"github.com/newrelic/release-toolkit/src/app/common"
	"github.com/newrelic/release-toolkit/src/app/gha"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/changelog/sources/dependabot"
	"github.com/newrelic/release-toolkit/src/changelog/sources/markdown"
	"github.com/newrelic/release-toolkit/src/changelog/sources/renovate"
	"github.com/newrelic/release-toolkit/src/git"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

const (
	markdownPathFlag  = "markdown"
	renovateFlag      = "renovate"
	dependabotFlag    = "dependabot"
	tagPrefixFlag     = "tag-prefix"
	gitRootFlag       = "git-root"
	includedDirsFlag  = "included-dirs"
	excludedDirsFlag  = "excluded-dirs"
	includedFilesFlag = "included-files"
	excludedFilesFlag = "excluded-files"
	exitCodeFlag      = "exit-code"
)

const emptyChangelogOutput = "empty-changelog"

// ErrNoSources is returned if Generate is invoked without any source enabled.
var ErrNoSources = errors.New("cannot generate changelog yaml without at least one source enabled")

// Cmd is the cli.Command object for the generate-yaml command.
//
//nolint:gochecknoglobals // We could overengineer this to avoid the global command but I don't think it's worth it.
var Cmd = &cli.Command{
	Name:  "generate-yaml",
	Usage: "Generates a machine-readable changelog.yaml file from multiple sources, including bot commits and the Unreleased section of CHANGELOG.md",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    markdownPathFlag,
			EnvVars: common.EnvFor(markdownPathFlag),
			Usage:   "Gather changelog entries from the specified file",
			Value:   "CHANGELOG.md",
		},
		&cli.BoolFlag{
			Name:    renovateFlag,
			EnvVars: common.EnvFor(renovateFlag),
			Usage:   "Gather changelog entries from renovate commits since last tag",
			Value:   true,
		},
		&cli.StringSliceFlag{
			Name:    includedDirsFlag,
			EnvVars: common.EnvFor(includedDirsFlag),
			Usage: `Only scan commits scoping at least one file in any of the following comma-separated directories, relative to repository root (--dir) ` +
				`(Paths may not start with "/" or contain ".." or "." tokens)`,
		},
		&cli.StringSliceFlag{
			Name:    excludedDirsFlag,
			EnvVars: common.EnvFor(excludedDirsFlag),
			Usage: `Exclude commits whose changes only impact files in specified dirs relative to repository root (--dir) (separated by comma) ` +
				`(Paths may not start with "/" or contain ".." or "." tokens)`,
		},
		&cli.StringSliceFlag{
			Name:    includedFilesFlag,
			EnvVars: common.EnvFor(includedFilesFlag),
			Usage: `Only scan commits scoping at least one file in any of the following comma-separated ones, relative to repository root (--dir) ` +
				`(Paths may not start or end with "/" or contain ".." or "." tokens)`,
		},
		&cli.StringSliceFlag{
			Name:    excludedFilesFlag,
			EnvVars: common.EnvFor(excludedFilesFlag),
			Usage: `Exclude commits whose changes only impact the specified files, path are relative to repository root (--dir) (separated by comma) ` +
				`(Paths may not start or end with "/" or contain ".." or "." tokens)`,
		},
		&cli.BoolFlag{
			Name:    dependabotFlag,
			EnvVars: common.EnvFor(dependabotFlag),
			Usage:   "Gather changelog entries from dependabot commits since last tag",
			Value:   true,
		},
		// Flags for tag sources.
		&cli.StringFlag{
			Name:    tagPrefixFlag,
			EnvVars: common.EnvFor(tagPrefixFlag),
			Usage:   "Find commits since latest tag matching this prefix.",
			Value:   "",
		},
		&cli.StringFlag{
			Name:    gitRootFlag,
			EnvVars: common.EnvFor(gitRootFlag),
			Usage:   "Path to the git repo to get commits and tags for.",
			Value:   "./",
		},
		&cli.IntFlag{
			Name:    exitCodeFlag,
			EnvVars: common.EnvFor(exitCodeFlag),
			Usage:   "Exit code if generated changelog yaml is empty",
			Value:   1,
		},
	},
	Action: Generate,
}

type appendDepSrc func([]changelog.Source, git.TagsVersionGetter, git.CommitsGetter) []changelog.Source

// Generate is a command that creates a changelog.yaml file.
//
//nolint:gocyclo,cyclop
func Generate(cCtx *cli.Context) error {
	gh := gha.NewFromCli(cCtx)

	yamlPath := cCtx.String(common.YAMLFlag)
	chFile, err := os.Create(yamlPath)
	if err != nil {
		return fmt.Errorf("opening changelog file %q: %w", yamlPath, err)
	}

	combinedChangelog := &changelog.Changelog{}
	sources := make([]changelog.Source, 0)

	if cCtx.Bool(renovateFlag) {
		appendDep := func(sources []changelog.Source, tgv git.TagsVersionGetter, getter git.CommitsGetter) []changelog.Source {
			return append(sources, renovate.NewSource(tgv, getter))
		}
		sources, err = addDepSource(cCtx, sources, appendDep)
		if err != nil {
			return fmt.Errorf("adding renovate source: %w", err)
		}
	}

	if cCtx.Bool(dependabotFlag) {
		appendDep := func(sources []changelog.Source, tgv git.TagsVersionGetter, getter git.CommitsGetter) []changelog.Source {
			return append(sources, dependabot.NewSource(tgv, getter))
		}
		sources, err = addDepSource(cCtx, sources, appendDep)
		if err != nil {
			return fmt.Errorf("adding dependabot source: %w", err)
		}
	}

	if mdPath := cCtx.String(markdownPathFlag); mdPath != "" {
		var mdFile *os.File
		mdFile, err = os.Open(mdPath)
		if err != nil {
			return fmt.Errorf("opening %q: %w", mdPath, err)
		}

		sources = append(sources, markdown.New(mdFile))
	}

	if len(sources) == 0 {
		return ErrNoSources
	}

	for _, source := range sources {
		var ch *changelog.Changelog
		ch, err = source.Changelog()
		if err != nil {
			return fmt.Errorf("gathering changelog from source: %w", err)
		}

		combinedChangelog.Merge(ch)
	}

	err = yaml.NewEncoder(chFile).Encode(combinedChangelog)
	if err != nil {
		return fmt.Errorf("writing changelog to %q: %w", yamlPath, err)
	}

	gh.SetOutput(emptyChangelogOutput, combinedChangelog.Empty())

	exitCode := cCtx.Int(exitCodeFlag)
	if combinedChangelog.Empty() && exitCode != 0 {
		return cli.Exit("changelog is empty", exitCode)
	}

	return nil
}

func addDepSource(cCtx *cli.Context, sources []changelog.Source, appendDep appendDepSrc) ([]changelog.Source, error) {
	tvg, err := tagVersionGetter(cCtx)
	if err != nil {
		return nil, err
	}

	includedDirs := sanitizeValue(cCtx.StringSlice(includedDirsFlag))
	excludedDirs := sanitizeValue(cCtx.StringSlice(excludedDirsFlag))
	includedFiles := sanitizeValue(cCtx.StringSlice(includedFilesFlag))
	excludedFiles := sanitizeValue(cCtx.StringSlice(excludedFilesFlag))

	gitCommitGetter := git.NewRepoCommitsGetter(cCtx.String(gitRootFlag))

	if len(includedDirs) > 0 || len(excludedDirs) > 0 || len(includedFiles) > 0 || len(excludedFiles) > 0 {
		commitFilter, err := git.NewCommitFilter(gitCommitGetter,
			git.IncludedDirs(includedDirs...),
			git.ExcludedDirs(excludedDirs...),
			git.IncludedFiles(includedFiles...),
			git.ExcludedFiles(excludedFiles...))
		if err != nil {
			return nil, fmt.Errorf("creating git commit filter: %w", err)
		}
		return appendDep(sources, tvg, commitFilter), nil
	}

	return appendDep(sources, tvg, gitCommitGetter), nil
}

func tagVersionGetter(cCtx *cli.Context) (*git.TagsSource, error) {
	workDir := cCtx.String(gitRootFlag)
	commitsGetter := git.NewRepoCommitsGetter(workDir)

	tagOpts := []git.TagOptionFunc{git.TagsMatchingCommits(commitsGetter)}
	if matching := cCtx.String(tagPrefixFlag); matching != "" {
		tagOpts = append(tagOpts, git.TagsMatchingRegex("^"+matching))
	}

	src, err := git.NewRepoTagsGetter(workDir, tagOpts...)
	if err != nil {
		return nil, fmt.Errorf("creating source for git tags: %w", err)
	}

	var versionOpts []git.TagSourceOptionFunc
	if matching := cCtx.String(tagPrefixFlag); matching != "" {
		versionOpts = append(versionOpts, git.TagSourceReplacing(matching, ""))
	}

	tvsrc := git.NewTagsSource(src, versionOpts...)
	if err != nil {
		return nil, fmt.Errorf("creating version source from git tag source: %w", err)
	}

	return tvsrc, nil
}

func sanitizeValue(in []string) []string {
	// Even if the user passes "an empty string it is considered as an element "",
	// and translated to "./" causing an unexpected behaviour.
	if len(in) == 1 {
		if in[0] == "" {
			return []string{}
		}
	}
	return in
}
