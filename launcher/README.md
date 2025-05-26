# Launcher

[![Go Report Card](https://goreportcard.com/badge/github.com/luskaner/ageLANServer/launcher)](https://goreportcard.com/report/github.com/luskaner/ageLANServer/launcher)

The launcher is a tool that allows you to launch the game to connect to the LAN server. It also handles configuring the
system and reverting that configuration upon exit.

## Minimum system Requirements

- Windows (no S edition/mode):
    - 10 on x86-64 (recommended).
    - 11 on ARM.
- Linux:
    - Kernel 2.6.23 on x86-64 (recommended).
    - Kernel 3.1 on ARM64.

**Note: If you allow it to handle the hosts file, local certificate, or an elevated custom game launcher, it will
require admin rights elevation.**

## Features

## Base

- Revert changes (if not properly reverted before).

## Server

- Generate a self-signed certificate.
- Start the server.
- Discover the server.
- Stop the server.

## Client (via [`bin\config`](../launcher-config/README.md))

- Isolated metadata directory (except AoE I).
- Smart modify the hosts file.
- Smart install of a self-signed certificate.

# Agent (via [`bin\agent`](../launcher-agent/README.md))

- Revert changes.
- Smart re-broadcast the battle server through other network interfaces apart from the most priority one.

## Command Line

CLI is available with similar options as the configuration. You can see the available options with
`launcher -h`.

## Configuration

The configuration options are available in the
`config.toml` ([windows](resources/windows/configs/config.toml)/[linux](resources/unix/configs/config.toml)) and [
`config.game.toml` ([windows](resources/windows/configs/config.game.toml)/[linux](resources/unix/configs/config.game.toml))
files. The files contain comments that
should help you understand the options.

## Exit Codes

* [Base codes](../common/errors.go).
* [Launcher shared codes](../launcher-common/errors.go).
* [Launcher codes](internal/errors.go).
