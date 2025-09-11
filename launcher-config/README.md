# Launcher Config

This executable makes and revert configuration changes and is executed by `launcher` or manually:

- Isolated metadata directory (except AoE I).
- Isolated profiles directory.
- Hosts file (via `config-admin`).
- Install of a self-signed certificate for the current user (only on Windows) or local (in this case via
  `config-admin`).

It is also responsible for managing the lifecycle and communicating with `config-admin-agent`.
Resides in `bin` subdirectory.

## Command Line

CLI is available. You can see the available options with
`config -h`.

You may run `revert -a -e <game>` (where game is either `age1`, `age2` or `age3`) to revert all changes (forced).

## Exit Codes

* [Base codes](../common/errors.go).
* [Launcher shared codes](../launcher-common/errors.go).
* [Config codes](internal/errors.go).
