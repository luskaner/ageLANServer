# Launcher Config

This executable makes and revert configuration changes and is executed by `launcher` or manually:

- Isolated metadata directory (except AoE I).
- Isolated profiles directory.
- Hosts file (via `config-admin`).
- Install of a self-signed certificate for the current user (only on Windows) or local (in this case via
  `config-admin`).

It is also responsible for managing the lifecycle and communicating with `config-admin-agent`.
Resides in `bin` subdirectory.

## Command Line

CLI is available. You can see the available options with
`config -h`.

You may run `revert -a -e <game>` (where game is either `age1`, `age2`, `age3`, `age4` or `athens`) to revert all
changes (forced).

## Console output

Messages are marked by outcome, and the marker degrades in three levels: emoji on a terminal with the
font (`✅ ❌ ⚠️ 🔹 ⏳`), single cell symbols on a console with a UTF-8 code page (`✔ ✖ ▲ • »`), and
bracketed tokens everywhere else (`[ OK ]`, `[FAIL]`, `[WARN]`, `[INFO]`, `[WAIT]`). A Windows 7
`cmd.exe`, a pipe and a redirected file all get the tokens, and no colour at all.

Only the marker is coloured; the body keeps the terminal's own foreground, so a line of green does not
read as an alarm.

Every message keeps the wording it has always had, so nothing that used to be visible is lost on an
old terminal. The file logs written through `internal.Logger.Buffer` are never decorated.

The text a message names is styled too, so a program no longer has to be quoted to be recognisable:
server, config-admin and config-admin-agent come out bold on a grey chip, paths and addresses dim,
identifiers like `config_setup_hosts` dim, and `--flags` bold dim. With colour off none of that is
emitted and the text is exactly what it says. See
[`launcher/README.md`](../launcher/README.md#what-is-styled-inside-a-line) for the table.


Values the renderer marks are not quoted: a component name, a path, an address or a flag
is styled, and quoting it on top says the same thing twice. The `zenity` dialog is the
exception, because a graphical window has no styling to fall back on. See
[`launcher/README.md`](../launcher/README.md#no-redundant-quotes).

Set `AGE_LANSERVER_OUTPUT=auto|color|ascii` to take control of the decision; an unknown value is
ignored. `NO_COLOR` disables colour wherever it is set. See
[`launcher/README.md`](../launcher/README.md#console-output) for the full table.

## Exit Codes

* [Base codes](../common/errors.go).
* [Launcher shared codes](../launcher-common/errors.go).
* [Own codes](internal/errors.go).
