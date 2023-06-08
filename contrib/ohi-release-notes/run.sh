#!/bin/bash

set -euo pipefail

RT_PKG="github.com/newrelic/release-toolkit@latest"
DICTIONARY_URL="https://raw.githubusercontent.com/newrelic/release-toolkit/v1/contrib/ohi-release-notes/rt-dictionary.yml"
EXCLUDED_DEPENDENCIES_MANIFEST_URL="https://raw.githubusercontent.com/newrelic/release-toolkit/v1/contrib/ohi-release-notes/excluded-dependencies.yml"
ARGS="$*"


# creating a temporary folder where to build the rt binary and cleaning after exiting.
TEMP_DIR=$(mktemp -dt release-toolkit-XXX)
function cleanup() {
    rm -rf $TEMP_DIR    || true
    rm "${GIT_ROOT}/CHANGELOG.md.bak" || true
    rm "${GIT_ROOT}/changelog.yaml"   || true
}
trap cleanup EXIT


# usage and help command. It is also an error exit in case flags are not correct.
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
   $0 - release toolkit wrapper to create release notes for OHIs

USAGE:
   $0 [options]

DESCRIPTION:
   Wrapper for release toolkit that runs commands needed to create the release notes for an OHI:
    * rt validate-markdown
    * rt generate-yaml
    * rt is-empty
    * rt is-held
    * rt link-dependencies
    * rt update-markdown (with the version calculated from the next-version command)
    * rt render-changelog (with the version calculated from the next-version command)

   At the end of the run, this command should output two files and a string:
    * CHANGELOG.md updated with the last changelog rendered.
    * A CHANGELOG.partial.md with the changes for this release only.
    * The version that was computed for this release.

OPTIONS:
   --git-root                       Run all the command using this path as root
   --verbose                        Adds verbose mode to this script.
   --help                           Show this help message and exits.
   --excluded-dirs                  Exclude commits whose changes only impact files in specified dirs relative to repository root. Defaults to ".github".
   --excluded-files                 Exclude commits whose changes only impact files in the list, paths relative to repository root. Defaults to "".
   --included-dirs                  Only scan commits scoping at least one file in any of the following comma-separated directories Defaults to "".
   --included-files                 Only scan commits scoping at least one file in the following comma-separated list. Defaults to "".
   --no-fail                        Do not fail even in the held toggle is active
   --dictionary                     Sets the link dependency dictionary file path. Default file located at "$DICTIONARY_URL" is used.
   --excluded-dependencies-manifest Sets the excluded dependencies manifest. Default file located at "$EXCLUDED_DEPENDENCIES_MANIFEST_URL".

EOM
    exit $ERRNO
}


# parsing flags
EXCLUDED_DIRECTORIES_FLAG="--excluded-dirs=.github"
EXCLUDED_FILES_FLAG=""
INCLUDED_DIRECTORIES_FLAG=""
INCLUDED_FILES_FLAG=""
IS_HELD_FAIL="--fail"
DICTIONARY=".github/rt-dictionary.yml"
GIT_ROOT="."
EXCLUDED_DEPENDENCIES_MANIFEST=".github/excluded-dependencies.yml"

while true; do
    if [ -z "${1:-}" ]; then
        break;
    else
        case "${1}" in
            -v | --verbose ) set -x; echo "Called with these arguments: $ARGS"; shift ;;
            -h | --help ) help ;;
            # Flags for all
            --git-root ) GIT_ROOT="$2"; shift 2 ;;
            # Flags for generate-yaml
            --excluded-dirs ) EXCLUDED_DIRECTORIES_FLAG="--excluded-dirs=$2"; shift 2 ;;
            --excluded-files ) EXCLUDED_FILES_FLAG="--excluded-files=$2"; shift 2 ;;
            --included-dirs ) INCLUDED_DIRECTORIES_FLAG="--included-dirs=$2"; shift 2 ;;
            --included-files ) INCLUDED_FILES_FLAG="--included-files=$2"; shift 2 ;;
            --excluded-dependencies-manifest ) EXCLUDED_DEPENDENCIES_MANIFEST="$2"; shift 2 ;;
            # Flags for is-held
            --no-fail ) IS_HELD_FAIL=""; shift ;;
            # Flags for link-dependencies
            --dictionary ) DICTIONARY="$2"; shift 2 ;;
            * ) help "option is not accepted/parsable: \"$1\"" ;;
        esac
    fi
done

# building rt
GOBIN="${TEMP_DIR}" go install ${RT_PKG}
RT_BIN="${TEMP_DIR}/release-toolkit"
if ! [ -x $RT_BIN ]; then
    help "rt binary is not executable: \"${RT_BIN}\""
fi

# fetch default dictionary by default
if ! [ -f "$DICTIONARY" ]; then
    DICTIONARY="${TEMP_DIR}/rt-dictionary.yml"
    curl -s -o "$DICTIONARY" "$DICTIONARY_URL"
fi

# fetch default excluded-dependencies-manifest by default
if ! [ -f "$EXCLUDED_DEPENDENCIES_MANIFEST" ]; then
    EXCLUDED_DEPENDENCIES_MANIFEST="${TEMP_DIR}/excluded-dependencies.yml"
    curl -s -o "$EXCLUDED_DEPENDENCIES_MANIFEST" "$EXCLUDED_DEPENDENCIES_MANIFEST_URL"
fi

EXCLUDED_DEPENDENCIES_MANIFEST_FLAG="--excluded-dependencies-manifest=${EXCLUDED_DEPENDENCIES_MANIFEST}"

(
    cd "${GIT_ROOT}"

    # checking that working directory is correctly set
    ${RT_BIN} validate-markdown


    # generating the changelog
    GENERATE_YAML_COMMAND="${RT_BIN} generate-yaml ${EXCLUDED_DIRECTORIES_FLAG} ${EXCLUDED_DEPENDENCIES_MANIFEST_FLAG} ${EXCLUDED_FILES_FLAG} ${INCLUDED_DIRECTORIES_FLAG} ${INCLUDED_FILES_FLAG}"
    echo "${GENERATE_YAML_COMMAND}"
    ${RT_BIN} generate-yaml "$EXCLUDED_DIRECTORIES_FLAG" "$EXCLUDED_DEPENDENCIES_MANIFEST_FLAG" "$EXCLUDED_FILES_FLAG" "$INCLUDED_DIRECTORIES_FLAG" "$INCLUDED_FILES_FLAG"
    ${RT_BIN} is-empty > /dev/null
    ${RT_BIN} is-held "${IS_HELD_FAIL}" > /dev/null
    if [ -f "$DICTIONARY" ]; then
        ${RT_BIN} link-dependencies --dictionary "$DICTIONARY"
    else
        ${RT_BIN} link-dependencies
    fi
    NEXT_VERSION="$(${RT_BIN} next-version)"
    ${RT_BIN} update-markdown --version "$NEXT_VERSION"
    ${RT_BIN} render-changelog --version "$NEXT_VERSION"

    echo "Release title should be: $(grep -E "^## " CHANGELOG.partial.md | sed 's|^## ||')"
)
