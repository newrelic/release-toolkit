[![New Relic Experimental header](https://github.com/newrelic/opensource-website/raw/master/src/images/categories/Experimental.png)](https://opensource.newrelic.com/oss-category/#new-relic-experimental)

# 🛠️ Generate YAML changelog action

An action to run the release toolkit generate yaml command. It builds a changelog.yaml file from multiple sources.

## Example Usage

Example generating a changelog yaml from all sources excluding commits whose changes only impact files in `charts` dir:
```yaml
- name: Test generate changelog yaml
  uses: newrelic/release-toolkit/generate-yaml@v1
  with:
    excluded-dirs: charts
```

## Parameters

All parameters are optional and match the ones used for the cli command flag, you can see the values and the defaults in [here](../README.md#generate-yaml))

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

