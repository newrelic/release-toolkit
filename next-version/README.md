# 🛠️ `next-version`

Compute next version according to changelog.yaml and Semver conventions. It will output the major version and the minor too, so it can be used to tag actions and docker images that do not need to be pinned to the whole version but to respect the interface that it has.

## Example Usage

Example generating the next version from a repo with version tags with last release tag `my-project-1.0.1:
```yaml
- name: Calculate next version
  uses: newrelic/release-toolkit/next-version@v1
  with:
    tag-prefix: my-project-
```

If for example there are breaking changes, the output will be `v2.0.0`, `v2.0`, and `v2`.

## Outputs

`next-version`: Returns Semver next version, with leading v. E.g.: `v3.4.1`
`next-version-major`: Returns only the major version, with leading v. E.g.: `v3`
`next-version-major-minor`: Returns major and minor version, with leading v. E.g.: `v3.4`

## Parameters

All parameters are optional and match the ones used for the cli command flag, you can see the values and the defaults in [here](../README.md#next-version))

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
