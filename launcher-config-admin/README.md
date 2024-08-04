# Launcher Config Admin
[![Go Report Card](https://goreportcard.com/badge/github.com/luskaner/aoe2DELanServer/launcher-config-admin)](https://goreportcard.com/report/github.com/luskaner/aoe2DELanServer/launcher-config-admin)

This executable makes and revert configuration changes which require admin privileges and is executed by `config`:

- Hosts file.
- Install of a self-signed certificate for the local Pc.

It is not meant to be run directly, only via `config`.
Resides in `bin` subdirectory.

## Command Line

CLI is available. You can see the available options with
`config-admin -h`.

## Exit Codes

* [Base codes](../common/errors.go).
* [Launcher shared codes](../launcher-common/errors.go).
* [Config Admin codes](internal/errors.go).
