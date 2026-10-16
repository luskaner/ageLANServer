# ageLANServer Configurator

The GUI loads and edits the official `config.toml` contract. It can also create a new configuration from the server template without maintaining a second template copy.

The load screen accepts files dragged from the operating system. Wails supplies the native path to Go, where the file is parsed and validated before it reaches the editor.

## Development

Run from this directory:

```text
wails3 dev -config ./build/config.yml -port 9245
```

Build the frontend and run the Go tests with:

```text
npm run build
go test ./...
```

## Creating a configuration

Use **Create New Configuration** to open `server/resources/config/config.toml` as an unsaved draft. Select the destination when saving. The configuration is validated in Go before it is created; an existing destination requires an explicit overwrite confirmation.

The template is searched in the installed `resources/config` directory and in the repository layout used during development. The official server template remains the source of truth.
