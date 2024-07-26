# Watcher

The watcher is run after the launcher starts the game and is responsible for reverting the configuration applied by it
(via `config`) and stopping the server if necessary after the game exits. It resides in `bin` subdirectory.

## Exit Codes

* [Base codes](/common/errors.go).
* [Launcher shared codes](/launcher-common/errors.go).
* [Watcher codes](internal/errors.go).