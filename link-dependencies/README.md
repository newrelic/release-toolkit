# 🛠️ `link-dependencies`

Attempts to add links to the original changelogs for dependency bumps in changelog.yaml.

The link is computed automatically when the dependency name is a full route, like `github.com/org/package`, or it's got from a [dictionary file](../README.md#dictionary-file) when present.

## Example Usage

Example with a dictionary:
```yaml
- name: Link dependencies
  uses: newrelic/release-toolkit/link-dependencies@v1
  with:
    dictionary: .github/dictionary.yaml
```

Dictionary file:

```yaml
dictionary:
  common-library: "https://github.com/newrelic/helm-charts/releases/tag/common-library-{{.To}}"
  golangci-lint: "https://github.com/golangci/golangci-lint/releases/tag/{{.To.Original}}"
```

## Parameters

All parameters are optional and match the ones used for the cli command flag, you can see the values and the defaults in [here](../README_CLI.md#link-dependencies))
