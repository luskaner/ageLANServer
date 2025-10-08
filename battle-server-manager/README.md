# Battle-Server-Manager

The battle-server-manager is a tool that allows you to launch and stop Online-like Battle
Servers.

## Minimum system Requirements

- Windows without S edition/mode (recommended):
    - 10 on x86-64.
    - 11 on ARM.
- Linux with kernel 3.2:
    - x86-64.
    - ARM64.

## Features

* Start a new Battle Server for a game.
* Stop (and clean up) one or more running Battle Servers for a given combination of region/game.
* Clean up non-running Battle Servers configurations for a given combination of region/game.

## Command Line

CLI is available. You can see the available commands and options by running
`battle-server-launcher -h`. Most options are mutually exclusive with the ones in configuration files.

## Configuration

The configuration options are available in the [
`config.game.toml`](resources/config.game.toml) files. The files contain comments
that
should help you understand the options.

## Exit Codes

* [Base codes](../common/errors.go).
* [Own codes](internal/errors.go).
