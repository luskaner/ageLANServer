# Launcher Config Admin

This executable makes and revert configuration changes which require admin privileges and is executed by `config`:

- Hosts file.
- Install of a self-signed certificate for the local Pc.

It is not meant to be run directly, only via `config`.
Resides in `bin` subdirectory.

## Command Line

CLI is available. You can see the available options with
`config-admin -h`.

## Console output

Messages are marked by outcome, in the same three levels as `config`: emoji (`✅ ❌ ⚠️ 🔹 ⏳`), single
cell symbols (`✔ ✖ ▲ • »`), or bracketed tokens (`[ OK ]`, `[FAIL]`, `[WARN]`, `[INFO]`, `[WAIT]`). An
elevated process may still be attached to the same console, so this follows the same rules as `config`,
and every message keeps the wording it has always had. Only the marker is coloured. The file logs are
never decorated.

The text a message names is styled too: config-admin-agent and config-admin come out bold on a grey
chip, paths and addresses dim, identifiers like `config_setup_hosts` dim. See
[`launcher/README.md`](../launcher/README.md#what-is-styled-inside-a-line) for the table.


Values the renderer marks are not quoted: a component name, a path, an address or a flag
is styled, and quoting it on top says the same thing twice. The `zenity` dialog is the
exception, because a graphical window has no styling to fall back on. See
[`launcher/README.md`](../launcher/README.md#no-redundant-quotes).

Set `AGE_LANSERVER_OUTPUT=auto|color|ascii` to take control of the decision. See
[`launcher/README.md`](../launcher/README.md#console-output) for the full table.

## Exit Codes

* [Base codes](../common/errors.go).
* [Launcher shared codes](../launcher-common/errors.go).
* [Own codes](internal/errors.go).
