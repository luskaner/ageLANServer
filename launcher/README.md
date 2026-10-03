# Battle-Server-Manager

The launcher is a tool that allows you to launch the game to connect to the LAN server. It also handles configuring the
system and reverting that configuration upon exit.

## Minimum system Requirements

- Windows without S edition/mode (recommended):
    - 7 on x86-64 (10 or higher recommended).
    - 11 on ARM.
- Linux with kernel 3.2 (5.10 or higher recommended):
    - x86-64 (recommended).
    - ARM64.
- macOS Monterey v12 (Sonoma v14 or higher recommended):
    - Intel (recommended).
    - Apple Silicon. Required for AoE II: DE native Steam version in Launcher.

**Note: If you allow it to handle the hosts file, local certificate, or an elevated custom game launcher, it will
require admin rights elevation.**

## Features

## Server

- Generate a self-signed certificate.
- Start the server.
- Discover the server.
- Stop the server.

## Battle-Server-Manager

- Start the Online-like Battle Server.
- Stop the Online-like Battle Server.

## Client (via [`bin\config`](../launcher-config/README.md))

- Isolated metadata directory (except AoE I).
- Isolated profiles directory.
- Smart modify the hosts file.
- Smart install of a self-signed certificate.
- Add certificate to the game's trusted store (except AoE I and AoE IV).

All possible client modifications are reverted upon the launcher's exit.

## Dialogs

- Ask the interactive questions (which server to use, and whether to start one) in a graphical window
  instead of in the console, when the system supports it.
- Always available in Windows and macOS. In Linux-like systems it requires `qarma`, `zenity` or `matedialog`
  to be installed.
- Always falls back to the console, so the launcher is never left without a way to ask. See `Config.Dialog`
  and the `-d`, `--dialog` flag.

## Command Line

CLI is available with similar options as the configuration. You can see the available options with
`launcher -h`. Some configuration options are only exclusive to the CLI and some to the configuration files.
The graphical dialogs are enabled or disabled with `-d`, `--dialog` (`auto`, `true` or `false`).

## Console output

The launcher adapts its console output to what the terminal can actually show, so the same run is
readable on a modern terminal and on a `cmd.exe` from Windows 7.

Three independent things are decided, because they are independent:

| | |
|---|---|
| **Colour** | ANSI escapes, downsampled to what the console supports. |
| **Glyphs** | three levels of degradation: bracketed tokens, single cell symbols, emoji. |
| **Width** | Long messages are wrapped to the console width, or left alone when the width is unknown. |

The markers, one per outcome, in the three levels:

| Outcome | Emoji | Unicode | ASCII |
|---|---|---|---|
| Success | `✅` | `✔` | `[ OK ]` |
| Failure | `❌` | `✖` | `[FAIL]` |
| Warning | `⚠️` | `▲` | `[WARN]` |
| Information | `🔹` | `•` | `[INFO]` |
| In progress | `⏳` | `»` | `[WAIT]` |
| Sub-step | `↳` | `↳` | tab (indentation) |

A sub-step cannot be drawn out of ASCII, so it degrades to what ASCII does have, which is
indentation: an indented line says "this belongs to the one above" as well as an arrow does.

Every marker is the same width inside its level, so the text after it starts in the same
column on every line whatever the outcome is.

What to expect where:

| Terminal | Level | Colour |
|---|---|---|
| Windows 7 `cmd.exe` (code page 437, no VT) | ASCII | none |
| Windows 10 `cmd.exe` on a code page the program cannot raise | ASCII | yes |
| Windows 10 or 11, code page raised to UTF-8 | Unicode | yes |
| Windows Terminal | Emoji | yes |
| Linux/macOS with a UTF-8 locale | Unicode, or Emoji on kitty, alacritty, wezterm… | yes |
| `TERM=dumb`, `NO_COLOR`, or `CI=true` | Emoji capped to Unicode | depends |
| Output redirected to a file or a pipe | ASCII | none |

The ASCII level is a guarantee, not a guess: every message keeps the exact wording it has always
had, and no byte outside printable ASCII (plus the tab) is emitted, so nothing can turn into a
replacement box.

On a modern Windows console whose code page could not be raised, the launcher prints one extra line
under `Configuration` naming the code page and the command that fixes it.

To take control of the decision:

- `--output auto|color|ascii` on the `launcher` (an explicit flag wins over the environment).
- `AGE_LANSERVER_OUTPUT=auto|color|ascii` works for `launcher`, `config` and `config-admin`.
  An unknown value is ignored rather than silently degrading the output.
- `NO_COLOR` disables colour wherever it is set, including when it is empty.
- `common/resources/start.bat` also runs `chcp 65001`, which is now redundant: the program sets the
  code page itself. It is harmless, and it keeps the window on UTF-8 for anything the user types after
  the launcher exits.

The file logs under `logs/` are never decorated: they stay plain text so `grep` keeps working.

### Shape of a run

The header is one line, the phases are headings, and nothing is boxed or tabulated:

```
  * launcher v1.4.0

Configuration
main config file: C:\Users\me\AppData\Roaming\AgeLANServer\config.toml
game config file: C:\Users\me\AppData\Roaming\AgeLANServer\config.age2.toml
game: age2

Initial teardown
» Cleaning up (if needed)...

Execution
» Setting up...
• Launching Age of Empires II...
✓ Game found on C:\Program Files (x86)\Steam\steam.exe
» Flushing cache...
✓ Successfully added host mappings
✗ Failed to start the server
      Received exit code: 1
```

There are two teardowns and they are named apart on purpose. The initial one undoes
what the previous run left behind and always happens; the final one undoes what this
run changed and appears only when there is something to undo, because a heading over
nothing tells the reader to wait for work that does not exist. Two headings called
the same thing looked like a bug.

A fact is announced once. The config files and the game are named in the summary and
nowhere else, because a message that appears twice is one more copy to keep in sync
and one more thing that looks wrong when the two copies disagree.

Nothing is aligned into columns. The summary is three lines of text, not a table, and
a table would need every key up front to line up and would put a column of whitespace
between the reader and the value on every line. A value too long for the window wraps
under itself.

Nothing is indented at the left margin either. The marker already separates a message
from the edge of the window, and every space in front of one is a space the first
word does not get. The hierarchy is carried by the marker, by `Detail` for subordinate
lines, and by the phase headings.

Phases are separated by a blank line and a heading rather than by a rule: a line of
dashes as wide as the window is a line the eye has to cross to reach the message under
it.
### What is styled inside a line

Only the marker is coloured, and the body keeps the terminal's own foreground. A whole
line of green reads as an alarm, and then the next green line looks like an error.

| | |
|---|---|
| **Component** — server, agent, launcher, config-admin, config-admin-agent, battle-server-manager | bold on a grey chip |
| **Path, address, URL** — `C:\Program Files\...\config.toml`, `192.168.1.50`, `https://…` | dim |
| **Identifier** — `Server.Executable`, `config_setup_hosts`, `serverStart` | dim |
| **Flag** — `--output` | bold dim |
| **Key** and **version** — the key of a summary row, the version in the header | `#6272A4`, `#888888` |
| **Marker** — success, failure, warning | `#50FA7B`, `#FF5555`, `#FFB86C` |
| **Marker** — information, in progress | `#8BE9FD`, `#6272A4` |
| **Phase heading** — Configuration, Initial teardown, Execution, Final teardown | `#8BE9FD`, bold |
| **Program name** — the one word at the top | `#FFFFFF`, bold |
| **Header glyph** — the `*` or `🚀` in front of it | `#6272A4` |

### No redundant quotes

A value the renderer already marks must not be quoted. Quoting it is a second,
uglier way of saying the same thing, and unlike the styling it survives into the
file log, where there is nothing to mark anything.

So: no quotes around a component name, a path, an address or a flag. `Running
Battle-Server`, `kill BattleServer.exe in task manager`, `with IP 192.168.8.1`.
A component is shown by the chip, an address by the dim styling, a flag by being
bold.

Two things keep them, and both are deliberate:

- the `zenity` dialog, which is a graphical window with no styling at all: there the
  quotes are the only thing that says "this is a program name".
- the two agent modules, which print plain text and never reach this renderer.

`TestNoQuotesAroundValuesTheRendererAlreadyMarks` enforces it by asking the styler
itself whether it would mark the token, so a module added later cannot reintroduce
the habit without the test saying so.

```
  ✔ Game found on C:\Program Files (x86)\Steam\steam.exe
  » Starting agent, authorize it in firewall if needed...
  • Communicating with config-admin-agent to add local cert and/or host mappings
```

A component name is only recognised on a token boundary, so `serverStart` and
`Config.Dialog` light up as identifiers and not as a `server` inside them.

### Long waits

Looking for the game, looking for a server and flushing the cache each take seconds and have nothing
to report while they run. Each one shows both of:

- an in place line that animates (`|`, `/`, `-`, `\`) and is then replaced by its outcome;
- the terminal's own progress indicator, via the [Windows Terminal progress bar
  sequences](https://learn.microsoft.com/en-us/windows/terminal/tutorials/progress-bar-sequences)
  (`OSC 9;4`), which fills the tab and the taskbar button and turns red or amber on a failure or a
  warning.

The indicator is only sent to a terminal known to understand it, identified by `WT_SESSION`. Anything
else would print the whole sequence at the cursor, which is why `--output ascii`, a pipe and a plain
conhost get no bar at all. It is always removed on the way out, including on a signal: a taskbar button
left half filled after the program is gone is worse than no button.

The in place animation has the same conservative gate: a console with a known width and colour.
Redirected or TierASCII output gets one plain line instead, because there is no cursor to move back
and a carriage return in a file is corruption.

The window title is never touched. `OSC 9;4` already draws its progress there, and overwriting the
title a user chose is not a decision a launcher should make.

### The console code page

A Windows console still boots into code page 437 or 850, and on those a UTF-8 symbol arrives as two
or three unrelated characters. Each program therefore puts its own console on UTF-8 at start up,
through `windows.SetConsoleOutputCP`, and settles the glyph tier after that. No wrapper script and no
manual `chcp` is needed.

It is deliberately narrow, because the code page is shared with every other process on the console:

- output only; the input code page, and therefore how the keyboard is decoded, is left alone;
- only when the standard output really is a console, so a piped run leaves the terminal that launched
  it untouched;
- only on Windows 10 build 19042 (Windows 11) or newer. Build 10586 was the first with VT and it
  shipped with the well known 65001 problems, where `cmd.exe` drew its own output wrong;
- never on Windows 7, which is the one case where the ASCII fallback is the contract;
- never with `--output ascii`, which is an explicit request for plain text;
- never in Windows Terminal, which is UTF-8 by construction.

If it cannot be set, the launcher prints one line under `Configuration` naming the code page and the
command that fixes it, because the fallback is the code page and not a preference, and without that
line the plain tokens look like the intended output rather than like what the console can draw.

Out of scope on purpose: any Bubble Tea TUI. See `TERMINAL.md`.

## Configuration

The configuration options are available in the [`config.toml`](resources/config.toml) and [
`config.game.toml`](resources/config.game.toml) files. The files contain comments
that
should help you understand the options.

## Exit Codes

* [Base codes](../common/errors.go).
* [Launcher shared codes](../launcher-common/errors.go).
* [Own codes](internal/errors.go).
