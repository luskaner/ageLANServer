# Launcher

[![Go Report Card](https://goreportcard.com/badge/github.com/luskaner/aoe2DELanServer/launcher)](https://goreportcard.com/report/github.com/luskaner/aoe2DELanServer/launcher)

The launcher is a tool that allows you to launch the game to connect to the LAN server. It also handles configuring the
system and reverting that configuration upon exit.

## Minimum system Requirements

- Windows (no S edition/mode):
  - 10 on x86-64 (recommended).
  - 11 on ARM.
- Linux: *recent* distribution with Steam on x86-64 using Steam Play (plus [Proton Experimental](https://github.com/ValveSoftware/Proton/wiki/Requirements)).

**Note: If you allow it to handle the hosts file, local certificate, or an elevated custom game launcher, it will require admin rights elevation.**

## Features

## Server

- Generate a self-signed certificate.
- Start the server.
- Discover the server.
- Stop the server.

## Client (via [`bin\config`](../launcher-config/README.md))

- Isolated metadata directory.
- Isolated profiles directory.
- Smart modify the hosts file.
- Smart install of a self-signed certificate.

All possible client modifications are reverted upon the launcher's exit.

## Command Line

CLI is available with similar options as the configuration. You can see the available options with
`launcher -h`.

## Configuration

The configuration options are available in the [`config.ini`](resources/config.ini) file. The file contains comments
that
should help you understand the options.

## Exit Codes

* [Base codes](../common/errors.go).
* [Launcher shared codes](../launcher-common/errors.go).
* [Launcher codes](internal/errors.go).
