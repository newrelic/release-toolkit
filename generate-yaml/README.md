# 🛠️ `generate-yaml`

An action to run the release toolkit `generate-yaml` command. It builds a `changelog.yaml` file from multiple sources.

## Example Usage

Example generating a changelog yaml from all sources excluding commits whose changes only impact files in `charts` dir:
```yaml
- name: Test generate changelog yaml
  uses: newrelic/release-toolkit/generate-yaml@v1
  with:
    excluded-dirs: charts
```

## Changes from `CHANGELOG.md`

`generate-yaml` parses `CHANGELOG.md` looking for entries written by maintainers under a L2 `## Unreleased` header.

`generate-yaml` expects L3 headers grouping changes by type. L3 headers containing the following words (case-insensitive) will be categorized as such:
- `breaking`
- `security`
- `enhancement`
- `bugfix`

Any other L3 header under `## Unreleased` will be added, including the header, to the raw notes section. This section is echoed verbatim by `render-markdown` and `update-markdown`.

These L3 headers should be composed exclusively of unordered lists. Each item of that list is taken as a single "change", and assigned to the type matching the header it is placed on. Text under a change type header (e.g. `Breaking changes`, which matches `breaking`) that is not a list item is ignored. Non-change type headers (e.g. `Important notice`, which does not match any change type) can contain any type of markdown construct.

The `CHANGELOG.md` can be validated with the `validate` command to spot wrong entries.

---

User modified CHANGELOG.md Example:
```md
# Changelog
All notable changes are documented in this file.

## Unreleased

Content that will be skipped

### ⚠️️ Breaking changes ⚠️
- Support has been removed

### Some Bugfixes
- A bugfix
- A second bugfix

### Important announcement (note)
This is a release note

### Another announcement
- this is an itemized release note

## [0.0.1] - 2022-09-20
### Added
- First version
```

Generated changelog.yaml:
```yaml
notes: |-
  ### Important announcement (note)
  This is a release note

  ### Another announcement
  - this is an itemized release note
changes:
  - type: breaking
    message: Support has been removed
  - type: bugfix
    message: A bugfix
  - type: bugfix
    message: A second bugfix
dependencies: []
```

### Held releases

Maintainers can include an L2 `## Held` header in the `CHANGELOG.md` file. This header must contain a paragraph below it indicating the reason why automated releases are being held.

`generate-changelog` will set the boolean `held` property to `true` in `changelog.yaml` if it founds such a header. This flag can be consumed later in the pipeline to check if an automated workflow should continue releasing.

## Bot sources
Dependabot and renovate changelog entries will be gathered (unless deactivated) from dependabot/renovate commits since last tag.
The release toolkit `generate-yaml` command will detect a renovate/dependabot commit based on the author and message of the commit.

The release toolkit will add those dependency entries, trying to extract the following (only the name is mandatory):
- The name of the dependency
- The old version
- The new version
- The PR and commit

Example:
```yaml
notes: ""
changes: []
dependencies:
  - name: github.com/newrelic/a-dependency
    from: "2"
    to: "3"
    meta:
      pr: "101"
      commit: 55c763d4920ca45d673d518f5448134b6b38091e
```
