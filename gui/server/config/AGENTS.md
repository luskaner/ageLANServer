# Configurator — agent notes

This Go + React (Wails v3) application edits the server's `config.toml` in place.
The server is the source of truth: the editor must read and write the same keys,
with the same defaults, that `server/internal/config.go` and
`server/internal/cmd/root.go` use.

## Configuration contract

The model in `models.go` mirrors `server/internal/config.go`. Do not invent keys.

| TOML key | Notes |
| --- | --- |
| `Log` | Root boolean. |
| `GeneratePlatformUserId` | Root boolean. Only valid with `Authentication = 'disabled'`. |
| `Authentication` | **Root key**, not under `[Internet]`. Values: `required`, `cached`, `adaptive`, `disabled`. Older documents that still hold `[Internet] Authentication` are kept in sync on save, but the root key is the one the server reads. |
| `[Sign] Enabled` | Re-sign catalogues with this server's key. |
| `[Sign] PlatformUserIdExceptions` | List of platform user ids excluded from signing. |
| `[Internet] Enabled` | Allow DNS check, authentication and public IP lookup. |
| `[Internet] IP` | `'auto'` or an IPv4 address. |
| `[Games] Enabled` | `age1`, `age2`, `age3`, `age4`, `athens`. |
| `[Games.<game>] Hosts` | IPv4 addresses the server binds for that game; must be unique among enabled games. |
| `[[Games.<game>.BattleServers]]` | One table per battle server. See below. |
| `[Announcement]` | `Enabled`, `Multicast`, `Port`, `MulticastGroup`. |
| `[CustomFiles]` | `Enabled`, `Folder`, `MaxFileSize`, `MaxTotalSize`, `Expiry`. |

### Battle servers

The documented shape is an array of tables, not an inline array:

```toml
[[Games.age2.BattleServers]]
Region = 'local'
Name = 'Local'          # only AoE II: DE uses it; written only when non-empty
IPv4 = 'auto'           # or a user-accessible IPv4 address
BsPort = 30003
WebSocketPort = 30004
OutOfBandPort = 30005   # not written for age1
```

`configurationDocument` rewrites every game through `rewriteBattleServers`:

- removes an old inline `BattleServers = [...]` key and every
  `[[Games.<game>.BattleServers]]` block;
- writes the canonical tables back only when the list has entries, so a save
  never appends empty `BattleServers` and a second save is a no-op;
- keeps the rest of the document (comments, unknown keys) untouched.

The add dialog (`frontend/src/components/BattleServerDialog.tsx`) handles a
single entry: `GameSection` appends it to the form list. A new entry starts with
the ports used in `server/BattleServers.md` for that game and is validated
locally (region required and unique, IPv4 `auto` or valid, ports in range)
before it reaches the form. The IPv4 field reuses the app's `AutoManualCard`:
automatic (`auto`) by default, manual with the four-octet `IPv4Input`. Battle
servers can be managed for a disabled game too: their configuration is valid
even if the server does not load it until the game is enabled.

### Game pages

Each game is its own sub-item under **Games** in the sidebar. The page title is
the game name and the header carries its enabled switch. The page has two
collapsible cards:

- **Hosts**: the right side lists the hosts separated by commas; expanding shows
  one IPv4 input per host with add and delete buttons.
- **Battle servers**: the right side lists the server regions separated by
  commas (the region is unique within the game); expanding lists each server's
  region/name, address and ports with a delete button. "Add battle server" opens
  the single-entry dialog. Its panel uses the opaque Win11-style surface from
  `style.css` (`.app-modal-*`), never the translucent card fill.

### Defaults

`serverDefaults()` replicates the defaults in `server/internal/cmd/root.go`:
`Authentication = 'disabled'`, `Sign.Enabled = false`, `Internet.Enabled = true`,
`Internet.IP = 'auto'`, `Announcement` enabled with port `31978` and group
`239.31.97.8`, `CustomFiles` enabled with `auto`/`32MB`/`256MB`/`12h`, and
`Hosts = ['0.0.0.0']` for every game.

## Validation

Validation lives in Go (`validateConfiguration`) and the frontend only renders
it through `ConfigService.ValidateConfiguration`. Errors block saving; warnings
do not. It mirrors what the server would do at startup:

- authentication values and the `GeneratePlatformUserId`/Internet combinations;
- supported games, and at least one host for an enabled game;
- hosts must resolve to IPv4 and not be shared by two enabled games (disabled
  games only warn);
- enabled `age4` and `athens` require at least one battle server (the server
  fails to start otherwise); `age1`/`age2`/`age3` may use LAN battle servers;
- battle server `Region` is required and unique within the game, `IPv4` is
  `auto` or an IPv4 address, and the ports are in range;
- announcement port/group, and custom file sizes/duration (sizes accept
  `B/KB/MB/GB`, durations are Go durations).

## Development

Run from `gui/server/config`:

```text
wails3 dev -config ./build/config.yml -port 9245
```

**Set `NODE_ENV=development` first.** With `NODE_ENV=production`, the
`npm install` run by the Wails build task omits devDependencies, `tsc`
disappears from `node_modules/.bin`, and the dev build fails.

The Vite config uses `server.watch.usePolling` because the project may live on a
network drive where native watchers do not deliver events. Without polling,
edits can leave the dev server serving a stale module.

Build/typecheck/tests:

```text
npx tsc --noEmit          # typecheck without rewriting dist
go test ./...
```

Avoid `npm run build` while a `wails3 dev` session is being exercised: writing
`frontend/dist` can trigger a full reload and discard the form state.

## Live testing with MCP

Start with the MCP build tag and a fixed port:

```powershell
$env:NODE_ENV='development'
$env:WAILS_MCP='1'
$env:WAILS_MCP_PORT='9099'
wails3 dev -config ./build/config.yml -port 9245
```

The app serves a stateless JSON-RPC MCP endpoint at `http://127.0.0.1:9099/mcp`
(tools: `app_info`, `windows_list`, `window_control`, `js_eval`, `dom_html`,
`dom_query`, `screenshot_dom`, `mouse_*`, `keyboard_*`, `call_bound_method`,
`emit_event`, `wait_for_event`).

On mount the app tries the default configuration paths (`openConfigFile`). When
none exists it stays on the load screen and does not open the native picker;
the picker only opens when the user selects a file explicitly. In an automated
session, load the file with `files-dropped` (or use the Select button).

Useful workflow for an end-to-end check:

1. `emit_event` `files-dropped` with `{ "files": ["C:\\path\\config.toml"] }`
   to load a copy through the real drag-and-drop path.
2. `js_eval` / `dom_query` to inspect the rendered state; use `mouse_click` with
   a CSS selector for elements below the fold (the tool scrolls them into view).
3. `keyboard_type` / `keyboard_press` for real text input.
4. Click Save, then read the file back and confirm the exact TOML shape.

`ConfigService` is bound as `changeme.ConfigService`; the generated `Call.ByID`
ids (e.g. `2068630133` for `OpenConfigFileAtPath`) can be used from `js_eval`
when a direct service call is needed.

## Things that are easy to get wrong

- `Authentication` is a **root** key. The server's koanf defaults use the root
  key and there is no migration for `[Internet] Authentication`.
- `BattleServers` is an **array of tables**. Adding an inline `BattleServers =`
  key to a file that already has `[[...BattleServers]]` is a duplicate-key error.
- An empty battle server list must remove the existing definitions, not write
  `BattleServers = []`.
- The editor must remount when another document is loaded (see `loadKey` in
  `App.tsx`): Mantine's `useForm` keeps its initial values for the life of the
  component.
- The unstable go-toml parser does not set `Raw` for table headers;
  `documentTables` finds header offsets by scanning whole header lines.
