# 🛠️ `render`

Renders a changelog.yaml as a markdown changelog section.

A new file `CHANGELOG.partial.md` will be created with the output of rendering changelog.yaml. 

## Example Usage

Example creating a `RELEASE-NOTES.md` file for a helm chart:
```yaml
- name: Test generate changelog yaml
  uses: newrelic/release-toolkit/render@v1
- name: Create chart release notes
  run: |
    mv ${GITHUB_WORKSPACE}/CHANGELOG.partial.md ${GITHUB_WORKSPACE}/charts/my-chart/RELEASE-NOTES.md
```

## Parameters

All parameters are optional and match the ones used for the cli command flag, you can see the values and the defaults in [here](../README.md#render))
