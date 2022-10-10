[![New Relic Experimental header](https://github.com/newrelic/opensource-website/raw/master/src/images/categories/Experimental.png)](https://opensource.newrelic.com/oss-category/#new-relic-experimental)

# 🛠️ release-toolkit

The release toolkit is a series of small, composable tools that interact together to allow building heavily flexible release pipelines.

This toolkit does not take care of pushing artifacts anywhere: It provides tools to decide whether a release can happen, and to automate the boring part of releasing, version bumping and changelogs.

The tools can be executed installing the [cli](#install) or using the different github actions described [here](#actions).

![release-toolkit.png](release-toolkit.png)
# Changelog YAML

The different release-toolkit components, either can be executed independently, or need to be fed with a custom-format yaml file from where all the current release changes, dependencies and notes will be computed.

This yaml file gets generated with the release toolkit command `generate-yaml` and has 3 changelog entry types:
- `notes`: Extra notes that don't imply a bump in the release version. 
- `changes`: Any changes in the project that cause a bump in the release version.
- `dependencies`: Dependency version bumps (they will imply a bump in the release version).

Example:

```yaml
notes: |-
  ### Important announcement (note)
  This is a release note
changes:
  - type: breaking
    message: Support has been removed
  - type: security
    message: Fixed a security issue
  - type: enhancement
    message: New feature has been added
dependencies:
  - name: docker/setup-buildx-action
    from: "1.0.0"
    to: "2.0.0"
    meta:
      pr: "100"
      commit: a72b98709dfa0d28cf7c73020f3dede670f7a37f
```

This `generate-yaml` command will extract this changes and notes from 3 different sources:

- User-facing changelog attached by the user in the PR.
- [Dependabot](https://github.com/dependabot) commits still not included in the last release.
- [Renovate](https://github.com/renovatebot/renovate) commits still not included in the last release.

## User-facing changelog

The release toolkit is intended to read entries from a CHANGELOG.md following the [keep a changelog format](https://keepachangelog.com/). 

Any change or note, the user wants to add to the next release, should be placed under the Level2 header `Unreleased`.

The `generate-yaml` command will read the entries in the `Unreleased` section, adding all the correct entries to the changelog.yaml file. We describe the types of entries in the next section.

### Changes

The generate-yaml command will only create entries in the changes section for content matching:
- Level3 Header **containing** one of the matching keywords in any letter case (table below).
- Body content must be an itemized list, or it won't be computed.

| Type          | Bump Type  | Example                   |
|---------------|------------|---------------------------|
| `breaking`    | `Major`    | `### Breaking change...`  | 
| `security`    | `Minor`    | `### This security...`    |
| `enhancement` | `Minor`    | `### enhancement...`      |
| `bugfix`      | `Patch`    | `### Bugfixes`            |

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

### Linked changelog
The link-dependencies command will try to attach a link to the bumped dependency version changelog for each dependency in the dependencies array.

To compute this link there the release toolkit tries to map the link executing this two tasks executed in the following order:
- Map the link from a [Dictionary file](#dictionary-file).
- [Auto-detected GitHub repository changelog](#automatically-detected-GitHub-repository-changelog).

#### Dictionary file
A dictionary is a YAML file with a root dictionary object, which contains a map from dependency names to a template that will be rendered into a URL pointing to its changelog.
The template link must be in Go tpl format and typically will include the {{.To.Original}} variable that will be replaced by the last bumped version.

Example dictionary file:
```yaml
dictionary:
  a-dependency: "https://github.com/my-org/a-dependency/releases/tag/{{.To.Original}}"
```

Example changelog.yaml:
```yaml
notes: ""
changes: []
dependencies:
  - name: a-dependency
    from: "2"
    to: "3"
    meta:
      pr: "101"
      commit: 55c763d4920ca45d673d518f5448134b6b38091e
```

When executing the `link-dependencies` command, the changelog.yaml will be modified to:
```yaml
notes: ""
changes: []
dependencies:
  - name: a-dependency
    from: "2"
    to: "3"
    meta:
      pr: "101"
      commit: 55c763d4920ca45d673d518f5448134b6b38091e
      changelog: https://github.com/my-org/a-dependency/releases/tag/3
```

* Using `{{.To.Original}}` instead of `{{.To}}` will preserve the leading `v` in case the version is `v*.*.*`.

#### Automatically detected GitHub repository changelog
When a dependency name is a full GitHub route, this pattern will be automatically used to create the link:
- `"https://github.com/{orgOrUser}/{repo-name}/releases/tag/{version}"`

Example changelog.yaml
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

When executing the `link-dependencies` command, the changelog.yaml will be modified to:
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
      changelog: https://github.com/my-org/a-dependency/releases/tag/3
```

If the automatic link generated is not correct it can be overwritten in the dictionary, since that task will be executed first.

## CLI

### Install
```shell
go install github.com/newrelic/release-toolkit@latest
```

### Generate YAML
Builds a changelog.yaml file from multiple sources.
```shell
rt generate-yaml [-flags]
```
| Flags           | Default        | Description                                                                                                                                                                                       |
|-----------------|----------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `markdown`      | `CHANGELOG.md` | Gather changelog entries from the specified file                                                                                                                                                  |
| `renovate`      | `true`         | Gather changelog entries from renovate commits since last tag                                                                                                                                     |
| `dependabot`    | `true`         | Gather changelog entries from dependabot commits since last tag                                                                                                                                   |
| `included-dirs` |                | Only scan commits scoping at least one file in any of the following comma-separated directories, relative to repository root (--dir) (Paths may not start with "/" or contain ".." or "." tokens) |
| `excluded-dirs` |                | Exclude commits whose changes only impact files in specified dirs relative to repository root (--dir) (separated by comma) (Paths may not start with "/" or contain ".." or "." tokens)           |
| `tag-prefix`    |                | Find commits since latest tag matching this prefix                                                                                                                                                |
| `git-root`      | `./`           | Path to the git repo to get commits and tags for                                                                                                                                                  |
| `exit-code`     | `1`            | Exit code if generated changelog is empty                                                                                                                                                         |                                                                                                                                             |

### Is held
```shell
rt is-held [-flags]
```
| Flags   | Default           | Description                                                            |
|---------|-------------------|------------------------------------------------------------------------|
| `yaml`  | `changelog.yaml`  | Path to the changelog.yaml file                                        |
| `fail`  | `false`           | If set, command will exit with a code of 1 if changelog should be held |
      
### Link dependencies
Add links to the original changelog for all the dependencies in a changelog.yml detecting the link if the name is a full route or getting the link from a dictionary file if present
```shell
rt link-dependencies [-flags]
```
| Flags           | Default          | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
|-----------------|------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `yaml`          | `changelog.yaml` | Path to the changelog.yaml file                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| `dictionary`    | `true`           | Path to a dictionary file mapping dependencies to their changelogs. A dictionary is a YAML file with a root dictionary object, which contains a map from dependency names to a template that will be rendered into a URL pointing to its changelog. The template link must be in Go tpl format and typically will include the {{.To.Original}} variable that will be replaced by the last bumped version (execute link-changelog with --sample flag to see a dictionary.yml sample)  |
| `sample`        |                  | Prints a sample dictionary to stdout                                                                                                                                                                                                                                                                                                                                                                                                                                                 |


### Next version
Current version is automatically discovered from git tags in the repository, in semver order.
Tags that do not conform to semver standards are ignored.
Several flags can be specified to limit the set of tags that are scanned, and to override both the current version being detected and the computed next version.
```shell
rt next-version [-flags]
```
| Flags          | Default          | Description                                                             |
|----------------|------------------|-------------------------------------------------------------------------|
| `yaml`         | `changelog.yaml` | Path to the changelog.yaml file                                         |
| `current`      |                  | If set, overrides current version autodetection and assumes this one    |
| `next`         |                  | If set, overrides next version computation and assumes this one instead |
| `git-root`     | `./`             | Path to the git repo to find tags on                                    |

### Render
Renders a changelog.yaml as a markdown changelog section.
```shell
rt render-markdown [-flags]
```
| Flags      | Default                | Description                                                                                                                    |
|------------|------------------------|--------------------------------------------------------------------------------------------------------------------------------|
| `yaml`     | `changelog.yaml`       | Path to the changelog.yaml file                                                                                                |
| `markdown` | `CHANGELOG.partial.md` | Path to the destination markdown file                                                                                          |
| `version`  |                        | Version to stamp in the changelog section header. If omitted, no version header will be generated                              |
| `date`     | `time.Now()`           | Date to stamp in the changelog section header, in YYYY-MM-DD format. If empty it will default to the current time (time.Now()) |                                                                                                                                                                                                          |

### Update markdown
Incorporates a changelog.yaml into a complete CHANGELOG.md.
```shell
rt update-markdown [-flags]
```
| Flags      | Default          | Description                                                                                                                    |
|------------|------------------|--------------------------------------------------------------------------------------------------------------------------------|
| `yaml`     | `changelog.yaml` | Path to the changelog.yaml file                                                                                                |
| `markdown` | `CHANGELOG.md`   | Path to the destination markdown file                                                                                          |
| `version`  |                  | Version to stamp in the changelog section header. If omitted, no version header will be generated                              |
| `date`     | `time.Now()`     | Date to stamp in the changelog section header, in YYYY-MM-DD format. If empty it will default to the current time (time.Now()) |                                                                                                                                                                                                          |

### Validate markdown
Prints errors if CHANGELOG.md has an invalid format.
```shell
rt validate-markdown [-flags]
```
| Flags           | Default        | Description                       |
|-----------------|----------------|-----------------------------------|
| `markdown`      | `CHANGELOG.md` | Validate specified changelog file |
| `exit-code`     | `1`            | Exit code when errors are found   |

## Actions

- [Generate YAML changelog](./generate-yaml/README.md)
- [Is Held](./is-held/README.md)
- [Link dependencies](./link-dependencies/README.md)
- [Next Version](./next-version/README.md)
- [render](./render/README.md)
- [Update markdown](./update-markdown/README.md)
- [Validate markdown](./validate-markdown/README.md)

## Contributing

Standard policy and procedure across the New Relic GitHub organization.

#### Useful Links
* [Code of Conduct](./CODE_OF_CONDUCT.md)
* [Security Policy](./SECURITY.md)
* [License](./LICENSE)
 
## Support

New Relic has open-sourced this project. This project is provided AS-IS WITHOUT WARRANTY OR DEDICATED SUPPORT. Issues and contributions should be reported to the project here on GitHub.

We encourage you to bring your experiences and questions to the [Explorers Hub](https://discuss.newrelic.com) where our community members collaborate on solutions and new ideas.

## License

release-toolkit is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.

## Disclaimer

This tool is provided by New Relic AS IS, without warranty of any kind. New Relic does not guarantee that the tool will: not cause any disruption to services or systems; provide results that are complete or 100% accurate; correct or cure any detected vulnerability; or provide specific remediation advice.

