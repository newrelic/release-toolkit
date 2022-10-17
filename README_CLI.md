# CLI

Release toolkit is also available as a CLI, which can be ran locally or in a CI/CD pipeline.

CLI commands automatically detect if they are running on GitHub actions. This autodetection can be overridden by explicitly passing `--gha=true` (or `false`), or by setting `GITHUB_ACTIONS` to `true` (or `false`).

## Install
```shell
go install github.com/newrelic/release-toolkit@latest
```

## Generate YAML
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

## Is held
```shell
rt is-held [-flags]
```
| Flags   | Default           | Description                                                            |
|---------|-------------------|------------------------------------------------------------------------|
| `yaml`  | `changelog.yaml`  | Path to the changelog.yaml file                                        |
| `fail`  | `false`           | If set, command will exit with a code of 1 if changelog should be held |

## Link dependencies
Add links to the original changelog for all the dependencies in a changelog.yml detecting the link if the name is a full route or getting the link from a dictionary file if present
```shell
rt link-dependencies [-flags]
```
| Flags           | Default          | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
|-----------------|------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `yaml`          | `changelog.yaml` | Path to the changelog.yaml file                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| `dictionary`    | `true`           | Path to a dictionary file mapping dependencies to their changelogs. A dictionary is a YAML file with a root dictionary object, which contains a map from dependency names to a template that will be rendered into a URL pointing to its changelog. The template link must be in Go tpl format and typically will include the {{.To.Original}} variable that will be replaced by the last bumped version (execute link-changelog with --sample flag to see a dictionary.yml sample)  |
| `sample`        |                  | Prints a sample dictionary to stdout                                                                                                                                                                                                                                                                                                                                                                                                                                                 |


## Next version
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

## Render
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

## Update markdown
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

## Validate markdown
Prints errors if CHANGELOG.md has an invalid format.
```shell
rt validate-markdown [-flags]
```
| Flags           | Default        | Description                       |
|-----------------|----------------|-----------------------------------|
| `markdown`      | `CHANGELOG.md` | Validate specified changelog file |
| `exit-code`     | `1`            | Exit code when errors are found   |
