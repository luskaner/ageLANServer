# Age LAN Server

Age LAN Server is a web server that allows you to play multiplayer **LAN** game modes without having an internet
connection **to the game server**  ensuring the game LAN functionality is still available even if the official
server
is in maintenance or are eventually shutdown.

Supported games:

* **Age of Empires: Definitive Edition**.
* **Age of Empires II: Definitive Edition**.
* **Age of Empires III: Definitive Edition**.

**You will still need a *custom launcher* to bypass the *online-only* restriction that is imposed by the game to being
connected to
the internet and Steam or Xbox Live, depending on the platform and version, to fully play offline.**

ℹ️ **Repository is now on maintenance only due to lack of contributions/funding and having completed the main goals.
Expect only updates to fix bugs and for dependencies/compiler.**

ℹ️ My other [project](https://github.com/luskaner/ageLANServerLauncherCompanion) provides the files and information
to download a Steam Emulator and play 100% offline.

*See more details
in [Questions and Answers (QA)](https://github.com/luskaner/ageLANServer/wiki/Questions-and-Answers-(QA))*.

## Features

- Scenarios.
- Map transferring in-lobby.
- Restore game.
- Data mods.
- Lobby chatting.
- Crossplay Steam & Xbox.

### Age of Empires II: Definitive Edition and Age of Empires III: Definitive Edition

<details>
<summary>List of features</summary>

- Co-Op Campaigns.
- Rematch.
- Invite player to lobby.
- Share lobby link.
- Player Search.

</details>

### Age of Empires III: Definitive Edition

<details>
<summary>List of features</summary>

- Channels.
- Whispering.

</details>

### Limitations

<details>
<summary>List of limitations</summary>

- Joining a game lobby from a link only works if the game is already running.
- Steam Workshop integration only works if using the official launcher.
- Lobbies can only be created in *LAN mode* which has its own limitations:
    - No Ranked.
    - No Spectate.
- Xbox and Steam friend integration.

</details>

#### Age of Empires II: Definitive Edition

<details>
<summary>List of limitations</summary>

- No Quick play.

</details>

#### Age of Empires III: Definitive Edition

<details>
<summary>List of limitations</summary>

- Friend list will instead show all online users as if they were friends.

</details>

## Unimplemented features

<details>
<summary>List of unimplemented features</summary>

- Achievements: only the official server should be able to. Meeting the requirements of an achievement during a
  match might cause issues (see [Troubleshooting](https://github.com/luskaner/ageLANServer/wiki/Troubleshooting)
  for more details).
- Changing player profile icon: the default will always be used.
- Leaderboards: will appear empty.
- Player stats: will appear empty.
- Clans: all players are without clans. Browsing clan will appear empty and creating one will always result in
  error.
- Lobby ban player: will appear like it works but doesn't.
- Report/Block player: will appear like it works but doesn't.

*Note: Most of these do not apply to Age of Empires: Definitive Edition.*

</details>

## Minimum system requirements

### Server

#### Stable

- Windows 10 (no S edition/mode).
- Windows Server 2016.
- Windows IoT (*except Arm32*).
- Linux: kernel 2.6.32 (see [here](https://go.dev/wiki/Linux) for more details).
- macOS: Big Sur (v11).

Admin rights or firewall permission to listen on port 443 (https) will likely be required depending on the operating
system.

<details>
<summary>Experimental</summary>

- BSD-based (OpenBSD, DragonFly BSD, FreeBSD and NetBSD).
- Solaris-based (Solaris and Illumos).
- AIX.

Note: For the full list see [minimum requirements for Go](https://go.dev/wiki/MinimumRequirements) 1.23.

</details>

### Launcher

- Windows (no S edition/mode):
    - 10 (1803 - Redstone 5) on x86-64 (recommended).
    - 11 on ARM.
- Linux: *recent* distribution with Steam on x86-64 using Steam Play.

**Note: If you allow it to handle the hosts file, local certificate, or an elevated custom game launcher, it will
require admin rights elevation.**

### Client

- Age of Empires: Definitive Edition
  on [Steam](https://store.steampowered.com/app/1017900/Age_of_Empires_Definitive_Edition)
  or [Xbox](https://www.xbox.com/games/store/age-of-empires-definitive-edition/9njwtjsvgvlj) (*only on
  Windows*). Requires version *100.2.31845.0* or later.
- Age of Empires II: Definitive Edition
  on [Steam](https://store.steampowered.com/app/813780/Age_of_Empires_II_Definitive_Edition)
  or [Xbox](https://www.xbox.com/games/store/age-of-empires-ii-definitive-edition/9N42SSSX2MTG/0010) (*only on
  Windows*). Requires a late 2023 version of the game (*use at least an early 2024 version*).
- Age of Empires III: Definitive Edition
  on [Steam](https://store.steampowered.com/app/933110/Age_of_Empires_III_Definitive_Edition)
  or [Xbox](https://www.xbox.com/games/store/age-of-empires-iii-definitive-edition/9n1hf804qxn4) (*only on
  Windows*). Requires a late 2023 version of the game (*use at least an early 2024 version*).

*Note: An up-to-date version is highly recommended as there are known issues with older versions.*
*Note 2: You could [compile](DEVELOPMENT.md) the programs with the domain changed to `aoe-api.reliclink.com`
in [here](common/common.go#L4) to support older versions but then newer versions would not work.*

## Binaries

See the [releases page](https://github.com/luskaner/ageLANServer/releases) for server and launcher binaries for a
subset of
supported operating systems.
<details>
    <summary>Provided archives</summary>

* Full:
    * Windows:
        * **10 on x86-64**: ..._full_*A.B.C*_win_x86-64.zip
        * **11 on ARM**: ..._full_*A.B.C*_win_arm64.tar.xz
    * Linux:
        * **x86-64**: ..._full_*A.B.C*_linux_x86-64.tar.xz
        * **ARM64**: ..._full_*A.B.C*_linux_arm64.tar.xz
* Launcher:
    * Windows:
        * **10 on x86-64**: ..._launcher_*A.B.C*_win_x86-64.zip
        * **11 on ARM**: ..._launcher_*A.B.C*_win_arm64.tar.xz
    * Linux:
        * **x86-64**: ..._launcher_*A.B.C*_linux_x86-64.tar.xz
        * **ARM64**: ..._launcher_*A.B.C*_linux_arm64.tar.xz
* Server:
    * Windows:
        * **10, Server 2025 or IoT on ARM64**: ..._server_*A.B.C*_win_arm64.zip
        * **10, Server 2016 or IoT on x86-64**: ..._server_*A.B.C*_win_x86-64.zip
        * **10 or 10 IoT on x86-32**: ..._server_*A.B.C*_win_x86-32.zip
    * Linux:
        * Kernel 3.1 on **ARM64**: ..._server_*A.B.C*_linux_arm64.tar.xz
        * Kernel 2.6.23 on **ARM32**:
            * ARMv5 (armel): ..._server_*A.B.C*_linux_arm-5.tar.gz
            * ARMv6 (sometimes called armhf): ..._server_*A.B.C*_linux_arm-6.tar.gz
        * Kernel 2.6.23 on **x86-64**: ..._server_*A.B.C*_linux_x86-64.tar.gz
        * Kernel 2.6.23 on **x86-32**: ..._server_*A.B.C*_linux_x86-32.tar.gz
    * macOS - Catalina (v10.15): ..._server_*A.B.C*_mac.tar.gz

</details>

*Note: If you are using Antivirus it may flag one or more executables as virus, this is a **false positive***.

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

Exceptions on tag/commit signature:

* Tags:
    * *v1.2.0-rc.5*: mantainer error.
* Commits:
    * *631cfa1* through *9eb66cf* (*both included*): rebase and merge PR issue.
    * *55697d4*: rebase of dependabot.

</details>

## Installation

Both the launcher and server work out of the box without any installation. Just download the archives,
decompress and run them.

## How it works

### Server

The server is simple web server that listens to the game's API requests. The server reimplements
the minimum required API surface to allow the game to work in LAN mode. NO data is stored or sent via the internet.

*Note: See the [server README](server/README.md) for more details.*

### Launcher

The launcher allows to easily play the game in LAN mode while still allowing the official launcher to be used for online
play.

<details>
    <summary>Features</summary>

- Automatically start/stop the server or connect to an existing one automatically.
- (Optional) Use an isolated metadata (except AoE I) and profile directories to avoid potential issues with the official
  game.
- (Optional) Modify the hosts file to:
    - Redirect the game's API requests to the LAN server.
    - Redirect the game CDN so it does not detect the official game status.
- (Optional) Install a self-signed certificate to allow the game to connect to the LAN server.
- Automatically find and start the game.

Afterwards, it reverses any changes to allow the official launcher to connect to the official servers.
</details>

*Note: See the [launcher README](launcher/README.md) for more details.*

## Simplest way to use it

1. **Download** the proper *full* asset from the latest
   stable release from https://github.com/luskaner/ageLANServer/releases.
2. **Uncompress** it somewhere.
3. If not using the Steam or Xbox launcher, **edit the `
   launcher/resources/config.<game>.toml` file** with a text editor (like Notepad)
   and modify
   the `Client.Executable` section to point to the game launcher path.
   **You will need to use a custom launcher (plus what my
   other [repo](https://github.com/luskaner/ageLANServerLauncherCompanion) provides) for 100% offline play**.
4. **Execute `launcher/launcher-<game>`: you will be asked for
   admin elevation and
   confirmation of other dialogs as
   needed, you
   will also need to allow the connections via the Microsoft Defender Firewall or any other.
5. **Repeat the above steps for every PC** you want to play in LAN with by running the *launcher*, the first PC to
   launch
   it will host the "server" and the rest will auto-discover and connect to it.
6. In the game, when hosting a new lobby, just make sure to set the server to **Use Local Lan Server** (AoE II),
   select **LAN** before creating the Lobby (AoE III) or select the "LAN" menu option (AoE I). In AoE I/II, setting it
   to
   public
   visibility is recommended.
7. If the lobby is Public, they can join directly in the browser or you can **Invite friends** by searching them by name
   and sending an invite as needed. You can share the link to join the lobby automatically (only works if already
   in-game).

## Separate server and launcher execution

<details>
    <summary>Server instructions</summary>

1. **Download** the proper *server* asset from latest stable release
   from https://github.com/luskaner/ageLANServer/releases.
2. **Generate the certificate** by simply executing `bin/genCert`.
3. If needed **edit the [config](server/resources/config/config.toml) file**.
4. **Run** the `server` binary for all games or the `server-` game-specific script.

</details>

<details>
    <summary>Launcher instructions</summary>

1. **Download** the proper *launcher* asset from latest stable release
   from https://github.com/luskaner/ageLANServer/releases.
3. If needed **edit the `launcher/resources/config.<game>.toml` file**. You will
   need to edit the
   `Client.Executable` section to point to the game launcher path if using a custom launcher which you will need to use
   a custom launcher for 100% offline play.
4. **Run** the `launcher-...`, Shell script.

*Note: If you have any issues run the `cleanup-<game>` Shell script.*

</details>

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md) to see how to develop and release builds.

## Terms of Use

You and all the clients connecting to your server are only authorized to use this software if:

- Owning a **legal license** of corresponding game (and all relevant DLC's).
- Comply with all the game terms of service.
- Use this software for personal use.
- Use this software in a LAN environment.

Disclaimer: This software is not affiliated or endorsed by any publisher or developer of the games.
