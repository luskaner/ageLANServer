# Launcher Agent

[![Go Report Card](https://goreportcard.com/badge/github.com/luskaner/ageLANServer/launcher-agent)](https://goreportcard.com/report/github.com/luskaner/ageLANServer/launcher-agent)

The agent is run after the launcher starts the game and is responsible for reverting the configuration applied by it
(via `config`) and stopping the server if necessary after the game exits. It may also apply a `BattleServer` fix if
configured to do so. It resides in `bin` subdirectory.

## Exit Codes

* [Base codes](../common/errors.go).
* [Launcher shared codes](../launcher-common/errors.go).
* [Agent codes](internal/errors.go).
