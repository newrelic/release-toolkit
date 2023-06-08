# 🛠️ `ohi-release-notes`

This is a wrapper of all the steps needed to update the changelog and render a snippet to be ready to be used as a release message. This contribution also includes a script that allow to replicate what this action does locally.

## Example Usage

```yaml
- uses: newrelic/release-toolkit/contrib/ohi-release-notes@v1
  id: release
- name: Commit updated changelog
  run: |
    git add CHANGELOG.md
    git commit -m "Update changelog with changes from ${{ steps.release.outputs.next-version }}"
    git push -u origin main
    gh release create ${{ steps.release.outputs.release-title }} --target $(git rev-parse HEAD) --notes-file CHANGELOG.partial.md
```

## Parameters

All parameters are optional:
  * `excluded-dirs` exclude commits whose changes only impact files in specified dirs relative to repository root. Defaults to ".github".
  * `excluded-files` Exclude commits whose changes only impact files in specified files relative to repository root. Defaults to "".
  * `included-dirs` Only scan commits scoping at least one file in any of the following comma-separated directories
  * `included-files` Only scan commits scoping at least one file in the following comma-separated list
  * `fail-if-held` fails if the held toggle is active
  * `dictionary` sets the link dependency dictionary file path. Defaults to ".github/rt-dictionary.yml".
  * `excluded-dependencies-manifest` sets the excluded dependencies manifest. Defaults to ".github/excluded-dependencies.yml".

## Outputs

  * `next-version` contains the calculated version for this.
  * `release-title` is the title of the release that includes `next-version` and the date it was done.
  * `release-changelog` contains the complete changelog of this release. Alias of the file `CHANGELOG.md`
  * `release-changelog-partial` contains the changelog for only this release. Alias of the file `CHANGELOG.partial.md`

This action also leaves the files `CHANGELOG.md` and `CHANGELOG.partial.md` at the working directory so they are also ready to be committed.

## Use script locally
There is a `run.sh` script that should do the same as this action: Leaves the files `CHANGELOG.md` and `CHANGELOG.partial.md` at the working directory and prints the title of the release with the next version and the date.

You can run it by bashpipeing this script:
```shell
curl "https://raw.githubusercontent.com/newrelic/release-toolkit/v1/contrib/ohi-release-notes/run.sh" | bash
```

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
