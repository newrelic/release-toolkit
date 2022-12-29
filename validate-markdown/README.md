[![New Relic Experimental header](https://github.com/newrelic/opensource-website/raw/master/src/images/categories/Experimental.png)](https://opensource.newrelic.com/oss-category/#new-relic-experimental)

# 🛠️ Validate markdown action

Validates a changelog in markdown format and prints errors if the changelog is invalid.

## Example Usage

Example exiting if errors are found:
```yaml
- name: Check if changelog is valid
  uses: newrelic/release-toolkit/validate-markdown@v1
```

Example without exiting:
```yaml
- name: Check if changelog is valid
  id: validate
  uses: newrelic/release-toolkit/validate-markdown@v1
  with:
    exit-code: 0
- run: |
    if [[ "${{ steps.validate.outputs.valid }}" != "true" ]]; then
      echo "markdown is not valid" >&2
      exit 1
    fi
```


#### Parameters

All parameters are optional and match the ones used for the cli command flag, you can see the values and the defaults in [here](../README.md#validate-markdown))

## Outputs

`valid`: Returns `true` if the changelog is valid

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
