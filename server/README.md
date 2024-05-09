# Server

The server module represents the core of the AoE2:DE LAN Server. It is a simple web server that listens to the game's API requests. The server reimplements the minimum required API surface to allow the game to work in LAN mode.

## System requirements

The server supports a very wide variety of operating systems and architectures. Basically any system you can compile [Go](https://go.dev/wiki/MinimumRequirements) 1.22 for. You can support even more operating systems by using an older Go version like 1.20 (few code changes might be required) that would enable Windows 7 or higher support, macOS High Sierra 10.13 or higher ... etc.

## Configuration

### Main

The few configuration options are available in the [`config.ini`](resources/config/config.ini) file. The file is self-explanatory and should be easy to understand.

### Login

The configuration file sent to the client upon login is [`login.json`](resources/config/login.json). Some options are easy to understand while others might require researching.

### Cloud

The game connects to a static cloud to download assets. The server is configured to replace the original calls to itself. The configuration file is [`cloudfilesIndex.json`](resources/config/cloudfilesIndex.json) and the corresponding files reside in the [`cloud`](resources/responses/cloud) directory.

### Other static responses

The server also serves some static responses for the game to work. The files are located in the [`responses`](resources/responses) base directory:

- [`Achievements`](resources/responses/achievements.json): List of achievements.
- [`Leaderboards`](resources/responses/leaderboards.json): List of leaderboards.
- [`Automatch maps`](resources/responses/automatchMaps.json): List of maps for automatch.
- [`Challenges`](resources/responses/challenges.json): List of challenges.
- [`Presence Data`](resources/responses/presenceData.json): Presence data. Basically if a player is online, offline or away.
- [`Item Definitions`](resources/responses/itemDefinitions.json): Definitions of items. Includes rewards, challenges and other items.
- [`Item Bundle Items`](resources/responses/itemBundleItems.json): Grouping of items into bundles.

*Note: These files might require updates to work with future game versions.*

## API endpoints

For documentation on how what each endpoints does, please refer to [LibreMatch documentation](https://wiki.librematch.org/rlink/game/start). Other endpoints are mostly self-explanatory.