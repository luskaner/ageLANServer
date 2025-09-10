# Battle Servers

The Battle Servers handle the logic of the matches themselves (and part of the lobby logic) including observing. They
are included in each
game installation. Battle Servers are independent of the main server and can hosted anywhere, including the same machine
as long as the server is configured properly.

## Should I use `LAN` or `Online-like` Battle Server?

| Feature               | LAN                                              | Online-like                                       |
|-----------------------|--------------------------------------------------|---------------------------------------------------|
| Execution             | $\color{Green}{\textsf{Automatic (game)}}$       | $\color{Yellow}{\textsf{Manual (Host)}}$          |
| Discovery             | $\color{Green}{\textsf{Automatic (game)}}$       | $\color{Yellow}{\textsf{Manual (Server)}}$        |
| Battle Server limit   | $\color{Yellow}{\textsf{1 per game}}$            | $\color{Green}{\textsf{None}}$                    |
| Latency               | $\color{Green}{\textsf{1-10ms}}$                 | $\color{Yellow}{\textsf{10-20ms}}$                |
| Observing             | $\color{Orange}{\textsf{Player observing only}}$ | $\color{Green}{\textsf{Full support}}$            | |
| Configurable          | $\color{Red}{\textsf{No}}$                       | $\color{Green}{\textsf{Yes, like ports..., etc}}$ |
| Can only host in self | $\color{Yellow}{\textsf{Yes}}$                   | $\color{Green}{\textsf{No}}$                      |

The LAN Battle Server is the easiest to use, but it has limitations. The Online-like Battle Server is more complex to
use, but it has more features and is more flexible.

*TL;DR: Use the LAN Battle Server unless you are having issues or need to use proper observing.*

## Starting the Battle Server

### Path

Depending on the game and version, the path can be one of these two:

* `Drive:\Path\To\Game\BattleServer.exe`
* `Drive:\Path\To\Game\BattleServer\BattleServer.exe`

### Executable

#### Arguments

##### Basic

* `-region` (string): the region identifier. It must be an unique string for a given game, but it is recommended to use
  a meaningful name such as the PC name in lower case. E.g., `roger-pc`.
* `-name` (string): the label for the region. It should be an unique string for a given game, but it is recommended to
  use a meaningful name such as the PC name. Only used in AoE II: DE. E.g., `Roger-PC`.

##### Ports

* `-relaybroadcastPort` (1-65535, UDP): used for broadcasting the server presence in the LAN. Recommended to be set to
  `0`
  to not interfere with the LAN Battle Server and because the server must be manually configured too.
* `-bsPort` (1-65535, TCP): main port to connect to the Battle Server.
* `-webSocketPort` (1-65535, TCP): websocket port to connect to the Battle Server.
* `-outOfBandPort` (1-65535, TCP): out-of-band (observer) communication. Missing in AoE: DE.

Note: All ports must be free and are recommended to be above `1024` to avoid requiring admin rights. The
host firewall will need to
allow incoming connections to these ports.

##### SSL

* `-sslCert` (string): path to the SSL certificate in PEM format. It can be any certificate that all clients trust.
  Recommended to use the server certificate as residing in `<server>/resources/certificates/cert.pem` and using the
  launcher on the clients to ensure it is trusted.
* `-sslKey` (string): path to the SSL private key in PEM format. Recommended to use the server key as residing in
  `<server>/resources/certificates/key.pem` and using the launcher on the clients to ensure the certificate is trusted.

### Examples

The examples assume the server is installed in `C:\AgeLANServer`, the server certificate and key are already
generated and that you are located in the battle server directory using the `Powershell` interpreter.

#### AoE: DE

`& ".\BattleServer.exe" -relayBroadcastPort 0 -region local -bsPort 30001 webSocketPort 30002 -sslKey C:\AgeLANServer\resources\certificates\key.pem -sslCert C:\AgeLANServer\resources\certificates\cert.pem`

#### AoE II: DE

`& ".\BattleServer.exe" -relayBroadcastPort 0 -region local -name Local -bsPort 30003 webSocketPort 30004 -outOfBandPort 30005 -sslKey C:\AgeLANServer\resources\certificates\key.pem -sslCert C:\AgeLANServer\resources\certificates\cert.pem`

#### AoE III: DE

`& ".\BattleServer.exe" -relayBroadcastPort 0 -region local -bsPort 30006 webSocketPort 30007 -outOfBandPort 30008 -sslKey C:\AgeLANServer\resources\certificates\key.pem -sslCert C:\AgeLANServer\resources\certificates\cert.pem`

## Configuring the Server

For each battle server instance you run you will need to edit `<server>/resources/config.toml` and inside `[Games]`
section, add `[[Games.<game>.BattleServers]]` where game is one of `age1`, `age3`, or `age3`.
inside that subsection the following properties are required (with the arguments matching):

| Battle Server argument | Server configuration key |
|------------------------|--------------------------|
| `-region`              | `Region`                 |
| `-name`                | `Name`                   |
| `-bsPort`              | `BsPort`                 |
| `-outOfBandPort`       | `OutOfBandPort`          |
| `-webSocketPort`       | `WebSocketPort`          |

Additionally, the `IPv4` which should point to a user-accessible IP. E.g., `192.168.1.2`.

### Examples

The examples mirror the battle server examples. They assume the battle servers are reachable at `192.168.1.2`

#### AoE: DE

```toml
[[Games.age1.BattleServers]]
Region = 'local'
IPv4 = '192.168.1.2'
BsPort = 30001
WebSocketPort = 30002
```

#### AoE II: DE

```toml
[[Games.age1.BattleServers]]
Region = 'local'
Name = 'Local'
IPv4 = '192.168.1.2'
BsPort = 30003
WebSocketPort = 30004
OutOfBandPort = 30005
```

#### AoE III: DE

```toml
[[Games.age1.BattleServers]]
Region = 'local'
IPv4 = '192.168.1.2'
BsPort = 30006
WebSocketPort = 30007
OutOfBandPort = 30008
```

## Simplest way to use an online-like Battle Server

### Assumptions

* You will only run a single Battle Server instance per game at most.
* The server is installed in `<server>`. E.g., `C:\AgeLANServer`.
* The game is installed in `<game>`. E.g. `C:\Program Files (x86)\Steam\steamapps\common\AoE2DE`.

Replace the placeholders with the actual paths.

### Steps

1. Make sure the server has the certificate pair generated in `<server>/resources/certificates`, otherwise, generate it
   by running `<server>/bin/genCert`.
2. Execute the `PowerShell` interpreter (can be installed on non-Windows systems too).
3. Change directory to the game directory with `cd "<game>"`.
4. Run `ls` to check if `BattleServer.exe` exists, if it does not, run `cd BattleServer`.
5. Time to run the `BattleServer.exe`, run one or more of the following commands depending on the game you want to;
    * AoE: DE:
   ```
   & ".\BattleServer.exe" -relayBroadcastPort 0 -region local -bsPort 30001 webSocketPort 30002 -sslKey "<server>/resources/certificates/key.pem" -sslCert "<server>/resources/certificates/cert.pem" 
   ```
    * AoE II: DE:
    ```
    & ".\BattleServer.exe" -relayBroadcastPort 0 -region local -name Local -bsPort 30003 webSocketPort 30004 -outOfBandPort 30005 -sslKey "<server>/resources/certificates/key.pem" -sslCert "<server>/resources/certificates/cert.pem"
    ```
    * AoE III: DE
    ```
    & ".\BattleServer.exe" -relayBroadcastPort 0 -region local -bsPort 30006 webSocketPort 30007 -outOfBandPort 30008 -sslKey "<server>/resources/certificates/key.pem" -sslCert "<server>/resources/certificates/cert.pem"
    ```
6. Open `<server>/resources/config.toml` and add the corresponding configuration inside `[Games]` for the game you are
   running the
   Battle Server for. Replace `<server-ip>` with the actual server IP address (e.g., `192.168.1.2`):
    * AoE: DE:
   ```toml
    [[Games.age1.BattleServers]]
    Region = 'local'
    IPv4 = '<server-ip>'
    BsPort = 30001
    WebSocketPort = 30002   
   ```
    * AoE II: DE:
    ```toml
    [[Games.age2.BattleServers]]
    Region = 'local'
    Name = 'Local'
    IPv4 = '<server-ip>'
    BsPort = 30003
    WebSocketPort = 30004
    OutOfBandPort = 30005
    ```
    * AoE III: DE:
    ```toml
    [[Games.age3.BattleServers]]
    Region = 'local'
    IPv4 = '<server-ip>'
    BsPort = 30006
    WebSocketPort = 30007
    OutOfBandPort = 30008
    ```
7. Start the game as you normally would and then:
    * AoE: DE: Click on "Multiplayer" then on "Create Game", you may select the "Region", "local", or leave as default.
      Players can
      join matches by clicking on "Multiplayer" and then on "Lobby Browser".
    * AoE II: DE: Click on "Multiplayer" then on "Host Game", you may select the "Server", "Local" or leave as default.
      Players can join matches by clicking on "Multiplayer" and then on "Find Custom Game" (Lobby Browser) or by being
      invited.
      Players can observe games by going to "Multiplayer" then "Find Custom Game" and finally "Spectate Games".
    * AoE III: DE: Click on "Multiplayer" (make sure the that "Online" is selected at top right) and then on "Host a
      Casual Game", you may select the "Region", "local", or leave as default. Players can join games by going to "
      Multiplayer" then "Browse Games" and, observe games by enabling here "Spectator Mode".

