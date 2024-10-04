# Age of Empires 2 Definitive Edition LAN Server

AoE2:DE LAN Server is a web server that allows you to play multiplayer **LAN** game modes without having an internet
connection **to the game server** paving the way to how the original AoE2 worked plus many features new to HD and DE
versions.

**You will still need a *custom launcher* to bypass the *online-only* restriction that is imposed by the game to being connected to
the internet and Steam or Xbox Live, depending on the platform and version, to fully play offline.**

*See more details in [Questions and Answers (QA)](https://github.com/luskaner/aoe2DELanServer/wiki/Questions-and-Answers-(QA))*.

## Features

- Co-Op Campaigns.
- Scenarios (including transferring the map):
    - Event Scenarios (will change depending on the server version and might require a update).
    - Custom Scenarios.
- ... all other game modes available by creating a lobby **only with server as "Use Local Lan Server"**.
- Rematch.
- Restore game.
- Data mods.
- Invite player to lobby.
- Share lobby link (joining it only works if the game is already running).
- Player Search.
- Chatting.
- Crossplay Steam & Xbox.

## Unsupported features
<details>
<summary>List of unsupported features</summary>
    
- Spectate games: Not compatible with Battle Server, would require a re-implementation.
- Not possible as it would require internet and some access to the user profile:
    - Steam & Xbox Friends.
- Not implemented:
    - Achievements: only the official server should be able to. Meeting the requirements of an achievement during a
      match might cause issues (see [QA](https://github.com/luskaner/aoe2DELanServer/wiki/Questions-and-Answers-(QA)) for more details).
    - Changing player profile icon: the default will always be used.
    - Leaderboards: will appear empty.
    - Player stats: will appear empty.
    - Clans: all players are without clans. Browsing clan will appear empty and creating one will always result in
      error.
    - Lobby ban player: will appear like it works but doesn't.
    - Report player: will appear like it works but doesn't.
    - Quick Play: no matchmaking is implemented, use official servers for this mode.
    - Ranked: no matchmaking is implemented, use official servers for this mode.

</details>

## System Requirements

### Server

#### Pre-built releases (Stable)

- Windows: 10 or higher (no S edition nor S mode), Server 2016 or higher. 32-bit/64-bit both Arm and x86.
- macOS: Catalina (v10.15) or higher.
- GNU/Linux: kernel 2.6.32 or higher (see [here](https://go.dev/wiki/Linux) for more details). 32-bit/64-bit both Arm
  and x86 (other archs are considered stable but they are not included).

Admin rights or firewall permission to listen to port 443 for https will likely be required (once or repeatedly)
depending on the operating system.

<details>
<summary>Buildable (Experimental)</summary>
    
- BSD-based (OpenBSD, DragonFly BSD, FreeBSD and NetBSD).
- Solaris-based (Solaris and Illumos).
- AIX.

Note: For the full list see [minimum requirements for Go](https://go.dev/wiki/MinimumRequirements) 1.22.
    
</details>

### Launcher

- GNU/Linux: recent SteamOS and Ubuntu (or derivatives) with official Steam installations on x86-64 are preferred.
  *arm64*
  variant is not currently supported by Steam but it is being worked on.
- Windows: 10 (no S edition nor S mode) or higher, (possibly Server 2016 or higher) all x86-64 (same as the game).
  Windows 11 on Arm (arm64) or higher, (possibly Server 2025 or higher) - no S mode - is also supported.

**Note: If you allow it to handle the hosts file, local certificate, or an elevated custom game launcher, it will
require
admin rights elevation.**

### Client

- Age of Empires 2 Definitive Edition on Steam (Microsoft Store - Xbox - is also supported on Windows).
- Up-to-date version of the game.

## Binaries

See the [releases page](https://github.com/luskaner/aoe2DELanServer/releases) for server and launcher binaries for a
subset of
supported operating systems.
<details>
    <summary>Provided archives</summary>
    
* Full:
    * Windows:
        * **10 (or higher) x86-64 (64 bits)**: aoe2DELanServer_full_*A.B.C*_win_x86-64.zip
        * **11 (or higher) on Arm (64 bits)**: aoe2DELanServer_full_*A.B.C*_win_arm64.tar.xz
    * GNU/Linux::
        * **x86-64**: aoe2DELanServer_full_*A.B.C*_linux_x86-64.tar.xz
        * **Arm64**: aoe2DELanServer_full_*A.B.C*_linux_arm64.tar.xz
* Launcher:
    * Windows:
        * **10 (or higher) x86-64 (64 bits)**: aoe2DELanServer_launcher_*A.B.C*_win_x86-64.zip
        * **11 (or higher) on Arm (64 bits)**: aoe2DELanServer_launcher_*A.B.C*_win_arm64.tar.xz
    * GNU/Linux:
        * **x86-64**: aoe2DELanServer_launcher_*A.B.C*_linux_x86-64.tar.xz
        * **Arm64**: aoe2DELanServer_launcher_*A.B.C*_linux_arm64.tar.xz
* Server:
    * Windows:
        * **10 (or higher) on Arm (64 bits)**: aoe2DELanServer_server_*A.B.C*_win_arm64.zip
        * **10 IoT on Arm (32 bits)**: aoe2DELanServer_server_*A.B.C*_win_arm32.zip
        * **10 (or higher) x86-64 (64 bits)**: aoe2DELanServer_server_*A.B.C*_win_x86-64.zip
        * **10 x86-32 (32 bits)**: aoe2DELanServer_server_*A.B.C*_win_x86-32.zip
    * GNU/Linux:
        * Kernel 3.1 or higher with **Arm64**: aoe2DELanServer_server_*A.B.C*_linux_arm64.tar.xz
        * Kernel 2.6.23 or higher with **Arm32**:
            * ARMv5 (armel): aoe2DELanServer_server_*A.B.C*_linux_arm-5.tar.gz
            * ARMv6 (sometimes called armhf): aoe2DELanServer_server_*A.B.C*_linux_arm-6.tar.gz
        * Kernel 2.6.23 or higher with **x86-64**: aoe2DELanServer_server_*A.B.C*_linux_x86-64.tar.gz
        * Kernel 2.6.23 or higher with **x86-32**: aoe2DELanServer_server_*A.B.C*_linux_x86-32.tar.gz
    * macOS - Catalina (v10.15) or higher: aoe2DELanServer_server_*A.B.C*_mac.tar.gz
  
</details>

*Note: If you are using Antivirus it may flag one or more executables as virus, this is
a **false positive***.

### Verification

The verification process ensures that the files you download are the same as the ones that were uploaded by the
maintainer.

<details>
    <summary>Verification steps</summary>

    1. Check the release tag is verified with the committer's signature key (*as all commits must be*).
    2. Download the ```..._checksums.txt``` and ```..._checksums.txt.sig``` files.
    3. Import the [release public key](release_public.key) and import it to your keyring if you haven't already.
    4. Verify the ```..._checksums.txt``` file with the ```..._checksums.txt.sig``` file.
    5. Verify the SHA-256 checksum list inside ```..._checksums.txt``` with the downloaded archives.

</details>

Exceptions on tag/commit signature:

* Tags:
    * *v1.2.0-rc.5*: mantainer error.
* Commits:
    * *631cfa1* through *9eb66cf* (*both included*): rebase and merge PR issue.
    * *55697d4*: rebase of dependabot.

## Installation

Both the launcher and server work out of the box without any installation. Just download the compressed archives,
decompress and run them.

## How it works

### Server

The server is simple web server that listens to the game's API requests. The server reimplements
the minimum required API surface to allow the game to work in LAN mode. It is completely safe as no data sent from the
client
is stored or sent to any other server.

*Note: See the [server README](server/README.md) for more details.*

### Launcher

The launcher allows to easily play the game in LAN mode while allowing the official launcher to be used for online play.

<details>
    <summary>Features</summary>
    
- Automatically start/stop the server or connect to an existing one automatically.
- (Optional) Use an isolated metadata and profile directories to avoid potential issues with the official game.
- (Optional) Modify the hosts file to
    - Redirect the game's API requests to the LAN server.
    - Redirect the game CDN so it does not detect the official game status.
- (Optional) Install a self-signed certificate to allow the game to connect to the LAN server.
- Automatically find and start the game.

Afterwards, it reverses any changes to allow the official launcher to connect to the official servers.
</details>

*Note: See the [launcher README](launcher/README.md) for more details.*

## Simplest way to use it

1. **Download** the proper *full* asset from the latest
   stable release from https://github.com/luskaner/aoe2DELanServer/releases.
2. **Uncompress** it somewhere.
3. If not using the Steam or Microsoft Store (Xbox) launcher, **edit the [launcher/config.ini](launcher/resources/config.ini)
   file**
   and modify
   the `Client.Executable` section to point to the game launcher path.
   You will need to use a custom launcher for 100% offline play.
4. **Execute `launcher/launcher`**: you will be asked for admin elevation and confirmation of other dialogs as
   needed, you
   will also need to allow the connections via the Microsoft Defender Firewall or any other.
5. **Repeat the above steps for every PC** you want to play in LAN with by running the `launcher`, the first PC to
   launch
   it will host the "server" and the rest will auto-discover and connect to it.
6. In the game, when hosting a new lobby, just make sure to set the server to **Use Local Lan Server**. Setting it to
   public
   visibility is recommended.
7. If the lobby is Public, they can join directly in the browser or you can **Invite friends** by searching them by name
   and sending an invite as needed. You can share the link to join the lobby automatically (only works if already
   in-game).

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md) to see how to develop and release builds.

## Terms of Use

You and all the clients connecting to your server are only authorized to use this software if:

- Owning a **legal license** of Age of Empires 2 Definitive Edition (and all relevant DLC's).
- Comply with all the game terms of service.
- Use this software for personal use.
- Use this software in a LAN environment.

Disclaimer: This software is not affiliated with Xbox Game Studios, Microsoft Corporation, Forgotten Empires LLC,
World's Edge LLC, or any other entity that is involved in the development of Age of Empires 2 Definitive Edition.
