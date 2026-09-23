# 🛠️ `release-notes-mdx`

Extracts a single version's section from `CHANGELOG.md` and renders it as an MDX file suitable for publishing on
NewRelic public docs site.

It picks the `## v<version> - <YYYY-MM-DD>` section from the changelog and writes YAML file in the format that the NewRelic public docs site expects for the release notes.

## Example Usage

```yaml
- name: Generate docs release notes
  uses: newrelic/release-toolkit/release-notes-mdx@v1
  with:
    subject: Agent Control
    version: 1.19.0
    repo: newrelic/newrelic-agent-control
```

## Parameters

All parameters match the ones used for the CLI command flags, you can see the values and the defaults
[here](../README_CLI.md#release-notes-mdx).

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
