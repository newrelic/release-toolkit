# 🛠️ `validate-changes`

Validates the YAML files, one per change, that hold the unreleased changelog entries (see [change files](../README.md#change-files-one-yaml-file-per-change)) and lists every problem found: empty files, missing or unknown `type`, missing or multiline `message`, unknown fields and invalid YAML.

## Example Usage

```yaml
- name: Check if the changelog change files are valid
  uses: newrelic/release-toolkit/validate-changes@v1
  with:
    changes-dir: .changelog
```

## Parameters

All parameters are optional:
  * `changes-dir` directory with the YAML files to validate. A directory that does not exist has no change files. Defaults to ".changelog".
  * `exit-code` exit code when errors are found. Defaults to "1".

## Outputs

`valid`: Returns `true` if all the change files are valid.

## Contributing

Standard policy and procedure across the New Relic GitHub organization.

#### Useful Links
* [Code of Conduct](../CODE_OF_CONDUCT.md)
* [Security Policy](../SECURITY.md)
* [License](../LICENSE)

## License

release-toolkit is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.
