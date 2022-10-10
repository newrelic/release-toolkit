[![New Relic Experimental header](https://github.com/newrelic/opensource-website/raw/master/src/images/categories/Experimental.png)](https://opensource.newrelic.com/oss-category/#new-relic-experimental)

# 🛠️ Is Held action

Outputs whether automated releases should be skipped.

## Example Usage

Example checking if changelog.yaml is held and accessing the output in next step:
```yaml
- name: Check if the release must be held
  id: held
  uses: newrelic/release-toolkit/is-held@v1
- run: |
    if [[ "${{ steps.held.outputs.is-held }}" == "true" ]]; then
      echo "release is-held" >&2
      exit 1
    fi
```

## Parameters

All parameters are optional and match the ones used for the cli command flag, you can see the values and the defaults in [here](../README.md#is-held))

## Outputs

`is-held`: Returns `true` if next release should not be automated

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

