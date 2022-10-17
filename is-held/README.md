# 🛠️ `is-held`

Outputs whether automated releases should be skipped.

## Example Usage

Example checking if changelog.yaml is held and accessing the output in next step:
```yaml
- name: Check if the release must be held
  id: held
  uses: newrelic/release-toolkit/is-held@v1
- run: |
    if [[ "${{ steps.held.outputs.is-held }}" == "true" ]]; then
      echo "Releases are being held, skipping weekly release" >&2
      exit 1
    fi
```

## Parameters

All parameters are optional and match the ones used for the cli command flag.

## Outputs

`is-held`: Returns `true` if next release should not be automated
