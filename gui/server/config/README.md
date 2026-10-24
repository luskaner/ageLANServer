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

## Release

The configurator is released as the portable `config` executable that ships with the Windows server archives.
It is built by the goreleaser configuration generated in `tools/scripts`, which builds the frontend and generates the
Windows resources (icon, manifest and version information) before compiling the application. The same happens in the
release workflow, so the archives carry the same binary as a local `task build`.

It is compiled against [luskaner/wails](https://github.com/luskaner/wails), a fork of Wails v3 that carries the
fixes for the two panics that made Wails unusable on 32-bit Windows. The fork is wired in through the `replace`
in the workspace (`go.work`) and it is republished, with the patch applied, on every new upstream v3 release.

## Creating a configuration

Use **Create New Configuration** to open `server/resources/config/config.toml` as an unsaved draft. Select the destination when saving. The configuration is validated in Go before it is created; an existing destination requires an explicit overwrite confirmation.

The template is searched in the installed `resources/config` directory and in the repository layout used during development. The official server template remains the source of truth.
