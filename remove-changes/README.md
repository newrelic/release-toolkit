# 🛠️ `remove-changes`

Removes the YAML files, one per change, that hold the unreleased changelog entries (see [change files](../README.md#change-files-one-yaml-file-per-change)), once they have been released. Only the files that `generate-yaml` reads are removed, so a template (`TEMPLATE.yaml`), a README or subdirectories are kept. Prints the path of each removed file.

## Example Usage

```yaml
- name: Remove the released change files
  uses: newrelic/release-toolkit/remove-changes@v1
  with:
    changes-dir: .changelog
```

## Parameters

All parameters are optional:
  * `changes-dir` directory with the YAML files to remove. Defaults to ".changelog".

## Contributing

Standard policy and procedure across the New Relic GitHub organization.

#### Useful Links
* [Code of Conduct](../CODE_OF_CONDUCT.md)
* [Security Policy](../SECURITY.md)
* [License](../LICENSE)

## License

release-toolkit is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.
