# Launcher Config Admin Agent

The launcher config admin agent is a service designed to avoid repeated admin elevation dialogs executing `config-admin`
while in the background. This agent is started/stopped by `config` as needed.

## Exit Codes

* [Base codes](/common/errors.go).
* [Launcher shared codes](/launcher-common/errors.go).
* [Config Admin codes](internal/errors.go) (some might only be sent through IPC).