# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

Unreleased section should follow [Release Toolkit](https://github.com/newrelic/release-toolkit#render-markdown-and-update-markdown)

## Unreleased

### Notes

`ohi-release-notes` action was only used by Coreint team and it has been moved to their repository: 
[`coreint-automation` PR](https://github.com/newrelic/coreint-automation/pull/83)

`generate-yaml` failed to create an empty YAML by default. The composable nature of `release-toolkit`
encourages the user to hack the YAML if needed. This is also needed so an empty YAML is there for the
other actions like `is-empty` or `is-held` to work properly.

`next-version` should also follow the composable nature of `release-toolkit`. But this part of the tool
should fail if there is no new version in case a user hack the YAML to a point that is not bumping the
version. Not failing can lead scripts to override an already existing version. This is also a change
on the default behavior.

### Breaking
- `ohi-release-notes` action has been moved to another repository
- `generate-yaml` does not fail to create an empty YAML by default
- `next-version` fail if there are no version to be bumped

## v1.2.0 - 2024-08-09

### 🚀 Enhancements
- Fix markdown validator to match entry-type

## v1.1.0 - 2024-04-09

### ⛓️ Dependencies
- Upgraded golang from 1.21.3-alpine to 1.22.0-alpine
- Upgraded golang.org/x/crypto from 0.0.0-20190701094942-4def268fd1a4 to 0.17.0

## v1.0.0 - 2023-11-02

### 🚀 Enhancements
- First release of release toolkit

