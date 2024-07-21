# Launcher

The launcher is a tool that allows you to launch the game to connect to the LAN server. It also handles configuring the
system and reverting that configuration upon exit.

**You will still need a way to bypass the *online-only* restriction that is imposed by the game to being connected to
Steam or Xbox Live depending on the version to fully play offline.**

## System Requirements

- Windows:
    * 10 (no S edition nor S mode) or higher, (possibly Server 2016 or higher) all x86-64 (same as the game).
    * 11 on Arm (no S mode) or higher (possibly Server 2025 or higher) all Arm64.
- Admin rights if:
    - You allow it to handle the hosts file.
    - You allow it to handle the local PC certificates.
    - If the game launcher requires it.

## Features

## Server

- Generate a self-signed certificate.
- Start the server.
- Discover the server.
- Stop the server.

## Client (via [`config`](/launcher-config/README.md))

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

* [Base codes](/common/errors.go).
* [Launcher shared codes](/launcher-common/errors.go).
* [Launcher codes](internal/errors.go).