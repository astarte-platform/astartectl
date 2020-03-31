# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [0.11.0-rc.1] - Unreleased
### Added
- Add `realm-management interfaces sync` subcommand

### Fixed
- appengine: Fix crash when retrieving nil values out of device interfaces
- appengine: Fix panic when passing appengine-url without realmmanagement-url (#73)
- appengine: data-snapshot should not fail entirely when an interface is not fetched from Realm Management

### Changed
- Moved the codebase to use astarte-go instead of the internal replicated tree
- `appengine devices get-samples` now handles aggregates with an explicit invocation rather than guessing
  it from the path

## [0.11.0-rc.0] - 2020-02-28
### Added
- Add appengine group subcommand
- Add support for credentials inhibition
- Add support for interfaces stats in device describe subcommand
- Add appengine stats subcommand
- Add support to database retention policy and database retention TTL

## [0.10.6] - 2020-02-27
### Added
- Added "cluster instance fetch-housekeeping-key", to fetch the Housekeeping Private key from a Cluster
- Added multiple API set support in utils gen-jwt

### Changed
- Removed explicit delimiters in default token regex: they were redundant

## [0.10.5] - 2020-01-25
### Added
- Added "cluster instance upgrade", to upgrade Astarte instances
- Added "cluster instance change-profile", to change an existing Astarte instance's deployment profile

### Fixed
- Fixed Cluster Resource parsing in some corner case situations
- Do not take into account prereleases when looking for latest versions

## [0.10.4] - 2019-12-11
### Added
- Added the new cluster command, to manage remote, Kubernetes-based, clusters
- pairing: add unregister subcommand, allowing to register again a device that already requested its
  credentials

### Fixed
- Avoid flaky parsing when "value" is a path token (#48)

## [0.10.3] - 2019-11-06
### Added
- appengine: Aggregates are correctly supported in get-samples
- appengine: get-samples performs path validation against the requested interface
- utils: add command to convert a Device ID to its UUID representation and viceversa
- Added common aliases for all commands where this is applicable
- appengine: devices data-snapshot now accepts an additional argument to print the snapshot of a specific interface only

### Changed
- "appengine device describe" has been renamed into "appengine device show"

## [0.10.2] - 2019-10-22
### Added
- appengine: Add an option to skip Realm Management checks, where possible
- Add possibility to use a token instead of the private key

## [0.10.1] - 2019-10-18
### Added
- Add commands to generate Device IDs and authentication JWTs
- Add gobuild.sh script
- Add shell completion generator command
- Add CI
- Add appengine command and subcommands

### Fixed
- Fix keypair generation
- Fix Datacenter Replication checks in realm creation command
- Fix a bug that prevented realm key to be set from the command line

### Changed
- utils gen-jwt accepts the private key through -k rather than through -p, just like all other commands

## [0.10.0] - 2019-09-20
### Added
- Initial release
