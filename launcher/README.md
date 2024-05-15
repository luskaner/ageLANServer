# Launcher 

The launcher is a tool that allows you to launch the game to connect to the LAN server. It also handles configuring the system and reverting that configuration upon exit.

**You will still need a way to bypass the *online-only* restriction that is imposed by the game to being connected to Steam or Xbox Live depending on the version to fully play offline.**

## System Requirements

- Windows: 10 or higher, (possibly Server 2016 or higher) all x86-64 (same as the game).
- If you allow it to handle the hosts file, it will require rights elevation.

## Features

## Server

- Generate a self-signed certificate.
- Start the server.
- Discover the server.
- Stop the server

## Client

- Isolated metadata directory.
- Isolated profiles directory.
- Smart modify the hosts file.
- Smart install of a self-signed certificate.

All possible client modifications are reverted upon the launcher's exit.

## Configuration

The configuration options are available in the [`config.ini`](config/config.ini) file. The file contains comments that should help you understand the options.
