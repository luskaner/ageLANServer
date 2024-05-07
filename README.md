# Age of Empires 2 Definitive Edition LAN Server

AoE2:DE LAN Server is a web server that allows you to play multiplayer **LAN** game modes without having an internet connection similar to how the original AoE2 worked plus many features new to HD and DE versions.

## Features

- Co-Op Campaigns.
- Scenarios (including transferring the map):
  - Event Scenarios*.
  - Custom Scenarios.
- ... all other game modes available by creating a lobby **only with server as "Use Local Lan Server"**.
- Rematch.
- Restore game.
- Data mods.
- Invite player to lobby (including via link).
- Player Search.
- Chatting (both in the lobby and in-game).
- Crossplay (cross-platform) Steam & Xbox (PC-only).

*\*Will change depending on the server version and might require an update.*

## Unsupported features

- Not compatible with Battle Server P2P (LAN):
  - Quick Play.
  - Ranked.
  - Spectate Games.
- Not possible as it would require internet and some access to the user profile:
  - Steam & Xbox Friends.
- Not implemented:
  - Changing player profile icon: the default will always be used.
  - Leaderboards: will appear empty.
  - Player stats: will appear empty.
  - Clans: all players are without clans. Browsing clan will appear empty and creating one will always result in error.
  - Lobby ban player: will appear like it works but doesn't.
  - Report player: will appear like it works but doesn't.

## System Requirements

### Server
- Windows: 10 or higher, Server 2016 or higher.
- MacOS: Catalina 10.15 or higher.
- GNU/Linux: *any supported distro, see the note below for details*.

Admin rights to listen to port 443 for https will likely be required (once or repeatedly) depending on the operating system.

Note: For the full list see [minimum requirements for Go](https://go.dev/wiki/MinimumRequirements) 1.22.

### Launcher
- Windows: 10 or higher, (possibly Server 2016 or higher) all x86-64 (same as the game).
- If you allow it to handle the hosts file, it will require rights elevation.

### Client
- Age of Empires 2 Definitive Edition - Steam or Microsoft Store.
- Up-to-date version of the game, however: 
  - As-is *should* work since around mid-to-late 2023.
  - Changing domain to *aoe-api.reliclick.com* should work for even older versions until around mid-to-late 2023.

## Binaries

See the [releases page](https://github.com/luskaner/aoe2DELanServer/releases) for server and launcher binaries for supported operating systems.

## Installation
Both the launcher and server work out of the box without any installation. Just download the compressed archives, decompress and run them.

## How it works

### Server
The server is simple web server that listens to the game's API requests. The server reimplements
the minimum required API surface to allow the game to work in LAN mode. It is completely safe as no data sent by the client
is stored or sent to any other server.

*Note: See the [server README](server/README.md) for more details.*

### Launcher

The launcher allows to easily play the game in LAN mode while allowing the official launcher to be used for online play. 

It can do the following setup steps for you:
- Automatically start/stop the server or connect to an existing one automatically.
- (Optional) Use an isolated metadata and profile directories to avoid potential issues with the official game.
- (Optional) Modify the hosts file to redirect the game's API requests to the LAN server.
- (Optional) Install a self-signed certificate to allow the game to connect to the LAN server. 
- Automatically find and start the game.

Afterwards, it reverses any changes to allow the official launcher to connect to the official servers.

*Note: See the [launcher README](launcher/README.md) for more details.*

## Terms of Use

You and all the clients connecting to your server are only authorized to use this software if:

* Owning a legal license of Age of Empires 2 Definitive Edition.
* Not using this software to cheat/hack and, in general, respect all the game terms of service.
* Use this software for personal use.
* Use this software in a LAN environment.

Disclaimer: This software is not affiliated with Microsoft Corporation, Forgotten Empires LLC, or any other entity that is involved in the development of Age of Empires 2 Definitive Edition.