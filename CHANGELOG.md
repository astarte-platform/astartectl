# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Add fish shell completions generator command

### Changed
- `realm-management interfaces {install,upgrade}` commands are run synchronously.

## [22.11.02] - 23/05/2023
### Changed
- `appengine device`: print a parametric command rather than a partial one with
  `--to-curl`when multiple API calls are involved (e.g. `send-data`).
- `appengine device send-data` returns a clear error when an interface is not found.
  See [#132](https://github.com/astarte-platform/astartectl/issues/132).

### Fixed
- `appengine data-snapshot` properly gathers and shows data snapshots from a device.
- context: do not warn when config is missing. Users will have to provide parameters by hand.
- `appengine get-samples` properly gathers and show samples from a device.

## [22.11.01] - 2023-03-15
### Added
- Add support for ignoring SSL errors while interacting with the Astarte APIs.
- Add the `--to-curl` flag to print a command-line equivalent of a command instead
  of running it.

### Changed
- `cluster instance deploy`: Astarte >= `v1.0.0` is deployed using the
  `api.astarte-platform.org/v1alpha2` API.
- Require at least Go 1.18 (due to astarte-go dep).
- Use Go 1.18 for releases.

### Fixed
- `appengine devices send-data` correctly parses integers as int32 instead of in64.
  Fix [#176](https://github.com/astarte-platform/astartectl/issues/176).

## [22.11.00] - 2022-12-06
### Added
- `cluster instances migrate storage-version` allows to migrate CRDs with `[v1alpha1, v1alpha2]`
  stored versions to just `[v1alpha2]`.

## [1.0.1] - 2022-12-05
## Added
- `cluster instance deploy`: add `--burst` flag to deploy a burst Astarte instance. It 
  should be used only in resource-constrained environments, such as CI. Only Astarte
  0.11.x and 1.0.x are supported.

### Fixed
- context: do not warn when config is missing. Users will have to provide parameters by hand.

## [1.0.0] - 2022-06-13
### Added
- `cluster instances migrate replace-voyager` allows to migrate a deprecated AstarteVoyagerIngress
  to an AstarteDefaultIngress resource.
- `utils show-jwt-claims` to display claims of an Astarte token.

### Changed
- `cluster show`: add operator-name and operator-namespace flags
- `utils interfaces validate` allows validating Astarte Interfaces

### Removed
- `cluster instance deploy`: remove profile choice, only deploy a basic Astarte instance.
- `cluster instance`: remove outdated `change-profile` subcommmand.
- `cluster instance`: remove outdated `upgrade` subcommmand.

## [1.0.0-beta.7] - 2022-02-09
### Added
- realm/interfaces/sync: Add non-interactive mode

## [1.0.0-beta.6] - 2022-01-13
### Added
- housekeeping/realms: Add non-interactive mode
- config: Allow querying for current cluster
- pairing/agent: Allow registering a device with a machine-friendly output

### Changed
- `utils gen-jwt` will now use the private keys specified in the context, if any, to
  generate the tokens without asking for a private key explicitly.

## [1.0.0-beta.5] - 2022-01-12
### Changed
- deploy: burst profile for 1.0 should not specify a set amount of CPU for Verne

## [1.0.0-beta.4] - 2022-01-12
### Added
- deploy: Support for VerneMQ SSL Listener (1.0.1+)
- deploy: Support for custom Broker port
- deploy: Support non-interactive scenarios

### Changed
- deploy: Allow up to 2 minutes for housekeeping to come up after deployment

## [1.0.0-beta.3] - 2022-01-04
### Added
- Support for 1.0 profiles

### Changed
- Generate new keypairs using elliptic curves instead of RSA.
- Updated Kubernetes APIs to 1.23
- Require at least Go 1.17 (due to Kubernetes deps)
- Use Go 1.17 for releases

## [1.0.0-beta.2] - 2021-03-26
### Changed
- Device metadata have been renamed to attributes.
- Default jwt expiration time to 8 hours (instead of 5 minutes).

## [1.0.0-beta.1] - 2021-02-16
### Added
- `cluster instance get-cluster-config` allows getting a cluster configuration out of the
  current cluster

### Changed
- astartectl configuration now works through a context system, kubectl-style
- `cluster instance deploy` now creates a new cluster config context upon successful cluster creation
- `housekeeping realms create` has completely different (and incompatible) semantics: it now allows supplying
  either a public, private or no key, and will create a new config context accordingly
- `cluster {install,upgrade,uninstall}-operator` commands are deprecated, print instructions to
  perform the same tasks using Helm.

## [0.11.3] - 2021-01-27
### Added
- Add support for printing device details in `appengine devices list`
- Add support for filtering recently active devices in `appengine devices list`
- Add support for filtering connected/disconnected devices in `appengine devices list`

## [0.11.2] - 2020-10-22
### Added
- Add support for EC keys for JWT generation.

### Changed
- Bump Go version requirement to >= 1.13.

## [0.11.1] - 2020-09-10
### Fixed
- appengine send-data: fix object aggregated with nested arrays handling

## [0.11.0] - 2020-04-14
### Fixed
- `appengine devices send-data` now parses integer values correctly for server-owned datastream aggregate
  interfaces
- `appengine devices data-snapshot` now handles partial failures in a better way without compromising the
  full results

## [0.11.0-rc.1] - 2020-04-01
### Added
- Add `realm-management interfaces sync` subcommand
- Add `appengine devices send-data` subcommand

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
