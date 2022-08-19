package generate

import (
	"errors"
	"fmt"
	"os"

	"github.com/newrelic/release-toolkit/app/common"
	"github.com/newrelic/release-toolkit/changelog"
	"github.com/newrelic/release-toolkit/changelog/sources/markdown"
	"github.com/newrelic/release-toolkit/dependabot"
	"github.com/newrelic/release-toolkit/git"
	"github.com/newrelic/release-toolkit/renovate"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

const (
	mdPathFlag     = "md"
	renovateFlag   = "renovate"
	dependabotFlag = "dependabot"
	tagPrefixFlag  = "tag-prefix"
	dirFlag        = "dir"
)

// ErrNoSources is returned if Generate is invoked without any source enabled.
var ErrNoSources = errors.New("cannot generate changelog without at least one source enabled")

// Cmd is the cli.Command object for the generate command.
//
//nolint:gochecknoglobals // We could overengineer this to avoid the global command but I don't think it's worth it.
var Cmd = &cli.Command{
	Name:  "generate",
	Usage: "Builds a changelog.yaml file from multiple sources",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    mdPathFlag,
			EnvVars: common.EnvFor(mdPathFlag),
			Usage:   "Gather changelog entries from the specified file",
			Value:   "CHANGELOG.md",
		},
		&cli.BoolFlag{
			Name:    renovateFlag,
			EnvVars: common.EnvFor(renovateFlag),
			Usage:   "Gather changelog entries from renovate commits since last tag",
			Value:   true,
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
			Name:    dirFlag,
			EnvVars: common.EnvFor(dirFlag),
			Usage:   "Path to the git repo to get commits and tags for.",
			Value:   "./",
		},
	},
	Action: Generate,
}

// Generate is a command that creates a changelog.yaml file.
//
//nolint:gocyclo,cyclop
func Generate(cCtx *cli.Context) error {
	yamlPath := cCtx.String(common.ChangelogFlag)
	chFile, err := os.Create(yamlPath)
	if err != nil {
		return fmt.Errorf("opening changelog file %q: %w", yamlPath, err)
	}

	combinedChangelog := &changelog.Changelog{}
	sources := make([]changelog.Source, 0)

	if cCtx.Bool(renovateFlag) {
		sources, err = addRenovate(cCtx, sources)
		if err != nil {
			return fmt.Errorf("adding renovate source: %w", err)
		}
	}

	if cCtx.Bool(dependabotFlag) {
		sources, err = addDependabot(cCtx, sources)
		if err != nil {
			return fmt.Errorf("adding dependabot source: %w", err)
		}
	}

	if mdPath := cCtx.String(mdPathFlag); mdPath != "" {
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

	return nil
}

func addRenovate(cCtx *cli.Context, sources []changelog.Source) ([]changelog.Source, error) {
	tvg, err := tagVersionGetter(cCtx)
	if err != nil {
		return nil, err
	}

	gitCommitGetter, err := git.NewRepoCommitsGetter(cCtx.String(dirFlag))
	if err != nil {
		return nil, fmt.Errorf("creating git commit getter: %w", err)
	}

	sources = append(sources, renovate.NewSource(tvg, gitCommitGetter))
	return sources, nil
}

func addDependabot(cCtx *cli.Context, sources []changelog.Source) ([]changelog.Source, error) {
	tvg, err := tagVersionGetter(cCtx)
	if err != nil {
		return nil, err
	}

	gitCommitGetter, err := git.NewRepoCommitsGetter(cCtx.String(dirFlag))
	if err != nil {
		return nil, fmt.Errorf("creating git commit getter: %w", err)
	}

	sources = append(sources, dependabot.NewSource(tvg, gitCommitGetter))
	return sources, nil
}

func tagVersionGetter(cCtx *cli.Context) (*git.TagsSource, error) {
	var tagOpts []git.TagOptionFunc
	if matching := cCtx.String(tagPrefixFlag); matching != "" {
		tagOpts = append(tagOpts, git.TagsMatching("^"+matching))
	}

	src, err := git.NewRepoTagsGetter(cCtx.String(dirFlag), tagOpts...)
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
