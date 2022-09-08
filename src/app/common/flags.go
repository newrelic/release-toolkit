package common

const (
	// YAMLFlag is the command line flag to specify the path to a changelog.yaml file.
	// This flag is common and used by most commands.
	YAMLFlag = "yaml"

	// GHAFlag is the flag used by commands to identify if they should output GHA-syntax to stdout.
	GHAFlag = "gha"
	// GHAEnv is the env var equivalent for GHAFlag
	// GITHUB_ACTIONS: Always set to true when GitHub Actions is running the workflow. You can use this
	// variable to differentiate when tests are being run locally or by GitHub Actions.
	// https://docs.github.com/en/actions/learn-github-actions/environment-variables#default-environment-variables
	GHAEnv = "GITHUB_ACTIONS"
)
