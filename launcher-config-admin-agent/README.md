# Launcher Config Admin Agent
[![Go Report Card](https://goreportcard.com/badge/github.com/luskaner/aoe2DELanServer/launcher-config-admin-agent)](https://goreportcard.com/report/github.com/luskaner/aoe2DELanServer/launcher-config-admin-agent)

The launcher config admin agent is a service designed to avoid repeated admin elevation dialogs executing `config-admin`
while in the background. This agent is started/stopped by `config` as needed.
Resides in `bin` subdirectory.

## Exit Codes

* [Base codes](../common/errors.go).
* [Launcher shared codes](../launcher-common/errors.go).
* [Config Admin codes](internal/errors.go) (some might only be sent through IPC).
