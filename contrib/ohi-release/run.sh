#!/bin/bash

set -euo pipefail

RT_PKG="github.com/newrelic/release-toolkit@latest"
ARGS="$*"


# creating a temporary folder where to build the rt binary and cleaning after exiting.
TEMP_BIN=$(mktemp -dt release-toolkit-XXX)
function cleanup() {
    rm -rf $TEMP_BIN    || true
    rm CHANGELOG.md.bak || true
    rm changelog.yaml   || true
}
trap cleanup EXIT


# usage and helm command. It is also an error exit in case flags are not correct.
function help() {
    set +x  # Disable verbosity. If it is enabled at this point, it is not needed anymore.
    ERRNO=0
    if ! [ -z "${1:-}" ]; then
        echo "ERROR:"
        echo "   $1"
        echo ""
        ERRNO=1
    fi

    cat <<EOM
NAME:
   $0 - release toolkit wrapper to release OHIs

USAGE:
   $0 [options]

DESCRIPTION:
   Wrapper for release toolkit that runs commands needed to release an OHI:
    * rt validate-markdown
    * rt generate-yaml
    * rt is-held
    * rt link-dependencies
    * rt update-markdown (with the version calculated from the next-version command)
    * rt render-changelog (with the version calculated from the next-version command)

   At the end of the run, this command should output three files:
    * CHANGELOG.md updated with the last changelog rendered (Old CHANGELOG backed up as CHANGELOG.md.bak)
    * A CHANGELOG.partial.md with the changes for this release only.

OPTIONS:
   --verbose        Adds verbose mode to this script.
   --help           Show this help message and exits.
   --excluded-dirs  Exclude commits whose changes only impact files in specified dirs relative to repository root. Defaults to ".github".
   --no-fail        Do not fail even in the held toggle is active
   --dictionary     Sets the link dependency dictionary. Defaults to ".github/rt-dictionary.yaml".

EOM
    exit $ERRNO
}


# parsing flags
EXCLUDED_DIRECTORIES=".github"
IS_HELD_FAIL="--fail"
DICTIONARY=".github/rt-dictionary.yaml"

while true; do
    if [ -z "${1:-}" ]; then
        break;
    else
        case "${1}" in
            -v | --verbose ) set -x; echo "Called with these arguments: $ARGS"; shift ;;
            -h | --help ) help ;;
            # Flags for generate-yaml
            --excluded-dirs ) EXCLUDED_DIRECTORIES="$2"; shift 2 ;;
            # Flags for is-held
            --no-fail ) IS_HELD_FAIL=""; shift ;;
            # Flags for link-dependencies
            --dictionary ) DICTIONARY="$2"; shift 2 ;;
            * ) help "option is not accepted/parsable: \"$1\"" ;;
        esac
    fi
done


# building rt
GOBIN="${TEMP_BIN}" go install ${RT_PKG}
RT_BIN="${TEMP_BIN}/release-toolkit"
if ! [ -x $RT_BIN ]; then
    help "rt binary is not executable: \"${RT_BIN}\""
fi


# checking that working directory is correctly set
${RT_BIN} validate-markdown


# generating the changelog
${RT_BIN} generate-yaml --excluded-dirs "$EXCLUDED_DIRECTORIES"
${RT_BIN} is-held ${IS_HELD_FAIL} > /dev/null
if [ -f "$DICTIONARY" ]; then
    ${RT_BIN} link-dependencies --dictionary "$DICTIONARY"
else
    ${RT_BIN} link-dependencies
fi
NEXT_VERSION="$(${RT_BIN} next-version)"
${RT_BIN} update-markdown --version "$NEXT_VERSION"
${RT_BIN} render-changelog --version "$NEXT_VERSION"

echo "Release title should be: $(grep -E "^## " CHANGELOG.partial.md | sed 's|^## ||')"
