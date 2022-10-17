# 🛠️ `update-markdown`

Incorporates the contents of changelog.yaml as a new version header in CHANGELOG.md.

## Example Usage

Example updating `CHANGELOG.md` with the new release having a new version header as `v1.2.3 - {now}`:
```yaml
- uses: newrelic/release-toolkit/update-markdown@v1
  with:
    version: "v1.2.3"
```

## Parameters

All parameters are optional and match the ones used for the cli command flag, you can see the values and the defaults in [here](../README.md#update-markdown))
