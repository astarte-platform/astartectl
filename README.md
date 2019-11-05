# astartectl

Command line utility to manage Astarte

## Installation

### Using homebrew

If you are using [homebrew](https://brew.sh/) on your system, you can install `astartectl` out of its tap:

```bash
brew tap astarte-platform/astarte
brew update
brew install astartectl
```

### Using binaries from Github Release

You can download latest `astartectl` binaries for all platforms from [GitHub Releases](https://github.com/astarte-platform/astartectl/releases).

### Using go get

Assuming you have you go installation and [GOPATH set up](https://github.com/golang/go/wiki/SettingGOPATH),
you can just run

`go get github.com/astarte-platform/astartectl`

The `astartectl` binary will be installed in your `GOPATH`.

## Configuration

If you are using `astartectl` a lot with the same deployment, it could be easier to use a configuration
file instead of using flags. By default `astartectl` will search for `.astartectl.yaml` in your home
directory, but you can make it use a custom configuration by setting its path in the `ASTARTECTL_CONFIG`
environment variable.

This is a minimal sample configuration

```yaml
# This will be used as base url and urls to housekeeping, realm-management and
# pairing will be build appending /<service-name> to the base url
url: https://<your API base url>
housekeeping:
  key: <path to your housekeeping key>
realm:
  name: <your realm name>
  key: <path to your realm key>
```

You can also use environment variables, by using the `ASTARTECTL` prefix and joining the configuration
key with `_`, (e.g. `ASTARTECTL_REALM_NAME`).

Flags always override configuration.

## Usage

Run `astartectl` to see available commands.
