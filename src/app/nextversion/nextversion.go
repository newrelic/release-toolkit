package nextversion

import (
	"errors"
	"fmt"
	"os"

	"github.com/Masterminds/semver"
	"github.com/newrelic/release-toolkit/src/app/common"
	"github.com/newrelic/release-toolkit/src/app/gha"
	"github.com/newrelic/release-toolkit/src/bump"
	"github.com/newrelic/release-toolkit/src/bumper"
	"github.com/newrelic/release-toolkit/src/changelog"
	"github.com/newrelic/release-toolkit/src/git"
	"github.com/newrelic/release-toolkit/src/version"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

const (
	outputPrefix      = "output-prefix"
	tagPrefix         = "tag-prefix"
	currentFlag       = "current"
	nextFlag          = "next"
	gitRootFlag       = "git-root"
	BumpCapFlag       = "bump-cap"
	DependencyCapFlag = "dependency-cap"
	failFlag          = "fail"
)

const (
	nextVersionOutput = "next-version"
	majorOutput       = "major"
	majorMinorOutput  = "major-minor"
)

// Cmd is the cli.Command object for the next-version command.
//
//nolint:gochecknoglobals // We could overengineer this to avoid the global command but I don't think it's worth it.
var Cmd = &cli.Command{
	Name:  "next-version",
	Usage: "Prints the next version according to the current one, the changelog.yaml file, and semver conventions.",
	UsageText: `Current version is automatically discovered from git tags in the repository, in semver order. 
Tags that do not conform to semver standards are ignored.
Several flags can be specified to limit the set of tags that are scanned, and to override both the current version being
detected and the computed next version.
next-version will exit with an error if no previous versions are found in the git repository.`,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    tagPrefix,
			EnvVars: common.EnvFor(tagPrefix),
			Usage:   "Consider only tags matching this prefix for figuring out the current version.",
			Value:   "",
		},
		&cli.StringFlag{
			Name:    outputPrefix,
			EnvVars: common.EnvFor(outputPrefix),
			Usage:   "The prefix to prepend when printing the output version.",
			Value:   "v",
		},
		&cli.StringFlag{
			Name:    currentFlag,
			EnvVars: common.EnvFor(currentFlag),
			Usage:   "If set, overrides current version autodetection and assumes this one.",
			Value:   "",
		},
		&cli.StringFlag{
			Name:    nextFlag,
			EnvVars: common.EnvFor(nextFlag),
			Usage:   "If set, overrides next version computation and assumes this one instead.",
			Value:   "",
		},
		&cli.StringFlag{
			Name:    gitRootFlag,
			EnvVars: common.EnvFor(gitRootFlag),
			Usage:   "Path to the git repo to find tags on.",
			Value:   "./",
		},
		&cli.StringFlag{
			Name:    BumpCapFlag,
			EnvVars: common.EnvFor(BumpCapFlag),
			Usage:   "In case of having to bump the version of the package, limit to this semVer type",
			Value:   string(bump.MajorName),
		},
		&cli.StringFlag{
			Name:    DependencyCapFlag,
			EnvVars: common.EnvFor(DependencyCapFlag),
			Usage:   "In case of having to bump the version of base on a dependency, limit to this semVer type",
			Value:   string(bump.MajorName),
		},
		&cli.BoolFlag{
			Name:    failFlag,
			EnvVars: common.EnvFor(failFlag),
			Usage: "If set, command will exit with a code of 1 if no new version bump is produced. If not set," +
				"the current version will be returned.",
			Value: false,
		},
	},
	Action: NextVersion,
}

// NextVersion is a command function which loads a changelog.yaml file from disk and computes what the next version
// should be according to semver standards.
//
//nolint:gocyclo,cyclop
func NextVersion(cCtx *cli.Context) error {
	gh := gha.NewFromCli(cCtx)

	nextOverride, err := parseNextFlag(cCtx.String(nextFlag))
	if err != nil {
		return err
	}

	chPath := cCtx.String(common.YAMLFlag)
	chFile, err := os.Open(chPath)
	if err != nil {
		return fmt.Errorf("opening changelog file %q: %w", chPath, err)
	}

	ch := changelog.Changelog{}
	err = yaml.NewDecoder(chFile).Decode(&ch)
	if err != nil {
		return fmt.Errorf("loading changelog from file: %w", err)
	}

	versionSrc, err := source(cCtx)
	if err != nil {
		return err
	}

	entryCap, err := bump.NameToType(cCtx.String(BumpCapFlag))
	if err != nil {
		return fmt.Errorf("parsing version bump cap: %w", err)
	}
	dependencyCap, err := bump.NameToType(cCtx.String(DependencyCapFlag))
	if err != nil {
		return fmt.Errorf("parsing dependency bump: %w", err)
	}

	bmpr := bumper.New(ch)
	bmpr.EntryCap = entryCap
	bmpr.DependencyCap = dependencyCap

	next, err := bmpr.BumpSource(versionSrc)

	// Other errors are computed after checking for overrides in the switch statement.
	if errors.Is(err, bumper.ErrEmptySource) {
		log.Errorf("Refusing to compute next version as no previous version was found. Please create an initial version first.")
		return fmt.Errorf("computing next version: %w", err)
	}

	switch {
	case nextOverride != nil && next != nil:
		if nextOverride.LessThan(next) {
			log.Warnf("Next version should be %v, overriding to lower version %v", next.String(), nextOverride.String())
		} else {
			log.Infof("Overriding next version from autocomputed %v to %v", next.String(), nextOverride.String())
		}
		next = nextOverride

	case nextOverride != nil && next == nil:
		log.Infof("Could not compute automatic bump, using overridden version")
		next = nextOverride

	case nextOverride == nil && next != nil:
		// If we don't have an override, the bumper did not bump anything, and the user has set `--fail`, then we should error out.
		if errors.Is(err, bumper.ErrNoNewVersion) {
			log.Warnf("None of the changelog entries produced a version bump, returning current version")
			if cCtx.Bool("fail") {
				return fmt.Errorf("failing by user request: %w", err)
			}
		}

	case nextOverride == nil && next == nil:
		return fmt.Errorf("bumping source: %w", err)
	}

	_, _ = fmt.Fprintf(cCtx.App.Writer, "%s\n", fmt.Sprintf("%s%s", cCtx.String(outputPrefix), next.String()))
	gh.SetOutput(nextVersionOutput, fmt.Sprintf("%s%s", cCtx.String(outputPrefix), next.String()))
	gh.SetOutput(majorOutput, fmt.Sprintf("%s%d", cCtx.String(outputPrefix), next.Major()))
	gh.SetOutput(majorMinorOutput, fmt.Sprintf("%s%d.%d", cCtx.String(outputPrefix), next.Major(), next.Minor()))

	return nil
}

//nolint:nilnil // A sentinel error would be better, but we don't bother as this fn is unexported and used only once.
func parseNextFlag(override string) (*semver.Version, error) {
	if override == "" {
		return nil, nil
	}

	next, err := semver.NewVersion(override)
	if err != nil {
		return nil, fmt.Errorf("parsing next version override: %w", err)
	}

	return next, nil
}

//nolint:ireturn,nolintlint // I do want to return an interface here.
func source(cCtx *cli.Context) (version.Source, error) {
	if override := cCtx.String(currentFlag); override != "" {
		return version.Static(override), nil
	}

	workDir := cCtx.String(gitRootFlag)
	getter := git.NewRepoCommitsGetter(workDir)

	tagOpts := []git.TagOptionFunc{git.TagsMatchingCommits(getter)}
	if prefix := cCtx.String(tagPrefix); prefix != "" {
		tagOpts = append(tagOpts, git.TagsMatchingRegex("^"+prefix))
	}

	tg, err := git.NewRepoTagsGetter(workDir, tagOpts...)
	if err != nil {
		return nil, fmt.Errorf("building repo tags lister: %w", err)
	}

	var srcOpts []git.TagSourceOptionFunc
	if prefix := cCtx.String(tagPrefix); prefix != "" {
		srcOpts = append(srcOpts, git.TagSourceReplacing(prefix, ""))
	}

	return git.NewTagsSource(tg, srcOpts...), nil
}
