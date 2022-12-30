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

## Contributing

Standard policy and procedure across the New Relic GitHub organization.

#### Useful Links
* [Code of Conduct](../CODE_OF_CONDUCT.md)
* [Security Policy](../SECURITY.md)
* [License](../LICENSE)

## Support

New Relic has open-sourced this project. This project is provided AS-IS WITHOUT WARRANTY OR DEDICATED SUPPORT. Issues and contributions should be reported to the project here on GitHub.

We encourage you to bring your experiences and questions to the [Explorers Hub](https://discuss.newrelic.com) where our community members collaborate on solutions and new ideas.

## License

release-toolkit is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.

## Disclaimer

This tool is provided by New Relic AS IS, without warranty of any kind. New Relic does not guarantee that the tool will: not cause any disruption to services or systems; provide results that are complete or 100% accurate; correct or cure any detected vulnerability; or provide specific remediation advice.

