[![New Relic Experimental header](https://github.com/newrelic/opensource-website/raw/master/src/images/categories/Experimental.png)](https://opensource.newrelic.com/oss-category/#new-relic-experimental)

# 🛠️ Generate YAML changelog action

An action to run the release toolkit generate yaml command. It builds a changelog.yaml file from multiple sources.

## Example Usage

Example generating a changelog yaml from all sources excluding commits whose changes only impact files in `charts` dir:
```yaml
- name: Test generate changelog yaml
  uses: newrelic/release-toolkit/generate-yaml@v1
  with:
    excluded-dirs: charts
```

## Parameters

### Changes

The generate-yaml command will only create entries in the changes section for content matching:
- Level3 Header **containing** one of the matching keywords in any letter case (table below).
- Body content must be an itemized list, or it won't be computed.

### Notes
Any other Level3 header not containing one of the key entries, will be added to the notes section and won't bump the version.

The Body in notes doesn't need to be itemized.

### Wrong data
Content skipped:
- Any content not child of a header will be skipped.
- Headers with level higher than 3 and their body.

Content invalid (generate will fail):
- Non itemized lists for changes entries.
- Headers with level less than 3.

The `CHANGELOG.md` can be validated with the [validate](#validate-markdown) command to spot wrong entries.

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

## Bot sources
Dependabot and renovate changelog entries will be gathered (unless deactivated) from dependabot/renovate commits since last tag.
The release toolkit `generate-yaml` command will detect a renovate/dependabot commit based on the author and message of the commit.

The release toolkit will add those dependency entries, trying to extract the following (only the name is mandatory):
- The name of the dependency
- The old version
- The new version
- The PR and commit
- [Linked changelog](#linked-changelog) (A link to the original repository release-notes for this release version, added with the link-dependencies command)

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
      changelog: https://github.com/newrelic/a-dependency/releases/tag/2
```
