# 🛠️ `next-version`

Compute next version according to changelog.yaml and Semver conventions.

## Example Usage

Example generating the next version from a repo with version tags with last release tag `my-project-1.0.1:
```yaml
- name: Calculate next version
  uses: newrelic/release-toolkit/next-version@v1
  with:
    tag-prefix: my-project-
```

If for example there are breaking changes, the output will be `v2.0.0`

## Outputs

`next-version`: Returns Semver next version, with leading v

## Parameters

All parameters are optional and match the ones used for the cli command flag, you can see the values and the defaults in [here](../README.md#next-version))
