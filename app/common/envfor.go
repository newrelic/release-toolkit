package common

import (
	"strings"
)

// EnvFor automatically calculates the environment var equivalent of a command line flag.
// It does so by uppercasing and replacing dashes `-` with underscores `_`.
// An optional list of prefixes without separator might be supplied, in which case they are prepended to the resulting
// env var in order.
func EnvFor(flag string, prefixes ...string) []string {
	env := flag

	allPrefixes := ""
	for _, prefix := range prefixes {
		allPrefixes += strings.ToUpper(prefix) + "_"
	}

	env = allPrefixes + env
	env = strings.ToUpper(env)
	env = strings.ReplaceAll(env, "-", "_")

	return []string{env}
}
