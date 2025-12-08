# Server

The server module represents the core of the LAN Server. It is a simple web server that listens to the game's
API requests. The server reimplements the minimum required API surface to allow the game to work in LAN mode.

## Minimum system requirements

#### Stable

- **Windows**: 10 (or equivalent, not Arm32).
- **Linux**: kernel 3.2 (see [here](https://go.dev/wiki/Linux) for more details).
- **macOS**: Big Sur (v11).

Admin rights or firewall permission to listen on port 443 (https) will likely be required depending on the operating
system.

<details>
<summary>Experimental</summary>

- BSD-based (OpenBSD, DragonFly BSD, FreeBSD and NetBSD).
- Solaris-based (Solaris and Illumos).
- AIX.

Note: For the full list see [minimum requirements for Go](https://go.dev/wiki/MinimumRequirements) 1.24.

</details>

## Domains

### Main

* `aoe-api.reliclink.com`: legacy domain for AoE I: DE, AoE II: DE and AoE III: DE.
* `*.worldsedgelink.com`: current domain for all games.

### Playfab

* `*.playfabapi.com`: currently used in AoM for Arena of the Gods content.

### Partially proxied

- `api.ageofempires.com`: except text moderation so it works normally even without internet.
- `cdn.ageofempires.com`: except server status so it always shows online.

Note proxied domains only override part of the official functionality but retain the rest of it.

## Configuration

### Certificate

The easiest way to generate a self-signed certificate is by running the ``bin/genCert`` executable (more
info [here](../server-genCert) or you may leave
that to
the ```launcher``` if you are hosting and running the launcher on same PC.

#### Self-signed certificate

The self signed certificate pair (``selfsigned_cert.pem`` and ``selfsigned_key.pem``) is generated specifically for
non-AoM games.

#### Default

The default certificate pair (``cert.pem`` and ``key.pem``) serves as the default and for AoM. It is signed by
`cacert.pem` certificate authority.

You can use your own certificate by (re)placing the `cert.pem` and `key.pem` files in the `resources/certificates`
directory.

### Main

The few configuration options are available in the [`config.toml`](resources/config/config.toml) file. The file is
self-explanatory and should be easy to understand.

### Login

The configuration file sent to the client upon login is `login.json` inside the game subdirectory in `resources/config`.
Some options
are
easy to understand while others might require researching.

### Cloud

The game connects to a static cloud to download assets. The server is configured to replace the original calls to
itself. The configuration file is `cloudfilesIndex.json` inside the game subdirectory in `resources/config` and the
corresponding
files reside in the [`cloud`](resources/responses/cloud) directory.

### Age of Empires III: Definitive Edition only

#### Chat Channels

The chat channels are defined in the [`chatChannels.json`](resources/config/age3/chatChannels.json) file.

### Other static responses

The server also serves some static responses for the game to work. The files are located in
the [`responses`](resources/responses) base directory.

#### Age of Empires: Definitive Edition

- [`Item Definitions`](resources/responses/age1/itemDefinitions.json): Definitions of items. Includes rewards,
  challenges and
  other items.

#### Age of Empires II: Definitive Edition

- [`Achievements`](resources/responses/age2/achievements.json): List of achievements.
- [`Leaderboards`](resources/responses/age2/leaderboards.json): List of leaderboards.
- [`Automatch maps`](resources/responses/age2/automatchMaps.json): List of maps for automatch.
- [`Challenges`](resources/responses/age2/challenges.json): List of challenges.
- [`Presence Data`](resources/responses/age2/presenceData.json): Presence data. Basically if a player is online, offline
  or
  away.
- [`Item Definitions`](resources/responses/age2/itemDefinitions.json): Definitions of items. Includes rewards,
  challenges and
  other items.
- [`Item Bundle Items`](resources/responses/age2/itemBundleItems.json): Grouping of items into bundles.

#### Age of Empires III: Definitive Edition

- [`Achievements`](resources/responses/age3/achievements.json): List of achievements.
- [`Leaderboards`](resources/responses/age3/leaderboards.json): List of leaderboards.
- [`Item Definitions`](resources/responses/age3/itemDefinitions.json): Definitions of items. Includes rewards,
  challenges and
  other items.

#### Age of Mythology: Retold

##### Main

- [`Achievements`](resources/responses/athens/achievements.json): List of achievements.
- [`Leaderboards`](resources/responses/athens/leaderboards.json): List of leaderboards.
- [`Challenges`](resources/responses/athens/challenges.json): List of challenges.
- [`Presence Data`](resources/responses/athens/presenceData.json): Presence data. Basically if a player is online,
  offline
  or
  away. Also includes the screen the player is in
- [`Item Definitions`](resources/responses/athens/itemDefinitions.json): Definitions of items. Includes rewards,
  challenges and
  other items.
- [`Item Bundle Items`](resources/responses/athens/itemBundleItems.json): Grouping of items into bundles.

##### Playfab

* [`public-production`](resources/responses/athens/playfab/public-production):
    * [`1`](resources/responses/athens/playfab/public-production/1): all files here are for testing purposes, or for
      pre-release game versions.
    * [`2`](resources/responses/athens/playfab/public-production/2): current version of the files:
        * [`manifest.json`](resources/responses/athens/playfab/public-production/2/manifest.json): Specifies this folder
          paths.
        * [`seasons/seasons0.json`](resources/responses/athens/playfab/public-production/2/seasons/season0.json):
          left-over when they were planning to add seasons to AotG story mode.
        * [`daily_skirmish_plus.json`](resources/responses/athens/playfab/public-production/2/daily_skirmish_plus.json):
          Specifies the settings and order for the Daily Celestial Challenges.
        * [`feature_flags.json`](resources/responses/athens/playfab/public-production/2/feature_flags.json): Feature
          flags, currently just disabling the beta tag for AotG.
        * [`feature_flags.json`](resources/responses/athens/playfab/public-production/2/known_blessings.json): Listing
          of all blessings, even blessings or levels not used. Used by the server to grant them all to the users.

## Starting and configuring Online-like Battle Servers

See [Battle Servers](./BattleServers.md) for how to set up and start manually-configured Online-like Battle-servers.
See [Battle Server Manager](../battle-server-manager) for how to do so automatically.

## Command Line

CLI is available with similar options as the configuration. You can see the available options with
`server -h`.

## API endpoints

For documentation on how what each endpoints does, please refer
to [LibreMatch documentation](https://wiki.librematch.org/rlink/game/start). Other endpoints are mostly
self-explanatory.

## Docker

See [Docker](../tools/server-docker) for information.

## Exit Codes

* [Base codes](../common/errors.go).
* [Own codes](internal/errors.go).
